package filesyetem_service

import (
	"ctb-cli/core"
	"ctb-cli/repositories"
	"ctb-cli/services/object_service"
	"fmt"
	"io/fs"
	"path/filepath"
)

// FileSystem implements the FileSystem interface
type FileSystem struct {
	objectService object_service.Service
	linkRepo      *repositories.LinkRepository
	keyService    core.KeyService

	openToWrite map[string]openToWrite
}

type openToWrite struct {
	id string
}

var _ core.FileSystemService = &FileSystem{}

// NewFileSystem creates a new instance of PersistFileSystem
func NewFileSystem(keyService core.KeyService, objectSerivce object_service.Service, linkRepository *repositories.LinkRepository) *FileSystem {
	fileSys := FileSystem{
		objectService: objectSerivce,
		linkRepo:      linkRepository,
		keyService:    keyService,
		openToWrite:   make(map[string]openToWrite),
	}

	return &fileSys
}

func (f *FileSystem) CreateDir(path string) error {
	err := f.CreateVaultInPath(path)
	if err != nil {
		return err
	}
	return f.linkRepo.CreateDir(path)
}

func (f *FileSystem) CreateVaultInPath(path string) error {
	parentKeyId := ""
	if filepath.Clean(path) != string(filepath.Separator) {
		parentPath := filepath.Dir(path)
		vaultLink, err := f.getVaultLink(parentPath)
		if err != nil {
			return err
		}
		parentKeyId = vaultLink.VaultId
	}
	vault, err := f.keyService.CreateVault(parentKeyId)
	if err != nil {
		return err
	}
	link := core.NewVaultLink(vault.Id, vault.KeyId)
	err = f.linkRepo.InsertVaultLink(path, link)
	if err != nil {
		return err
	}

	return nil
}

func (f *FileSystem) RemovePath(path string) (err error) {
	return f.linkRepo.Remove(path)
}

func (f *FileSystem) GetSubFiles(path string) (res []fs.FileInfo, err error) {
	subFiles, err := f.linkRepo.GetSubFiles(path)
	if err != nil {
		return nil, err
	}
	var infos []fs.FileInfo
	for _, subFile := range subFiles {
		if subFile.IsDir() {
			var info fs.FileInfo = FileInfo{
				isDir: true,
				name:  subFile.Name(),
			}
			infos = append(infos, info)
			continue
		} else {
			p := filepath.Join(path, subFile.Name())
			link, err := f.linkRepo.GetByPath(p)
			if err != nil {
				return nil, fmt.Errorf("error reading file size: %v", err)
			}
			var info fs.FileInfo = FileInfo{
				isDir: false,
				name:  subFile.Name(),
				size:  link.Size,
			}
			infos = append(infos, info)
		}

	}
	return infos, nil
}

func (f *FileSystem) RemoveDir(path string) (err error) {
	return f.linkRepo.RemoveDir(path)
}

func (f *FileSystem) CreateFile(path string) (err error) {
	id, err := core.NewUid()
	if err != nil {
		return err
	}
	_ = f.linkRepo.Create(path, core.Link{
		ObjectId: id,
		Size:     0,
	})
	err = f.objectService.Create(id)
	if err != nil {
		return err
	}
	f.openToWrite[path] = openToWrite{id: id}
	return
}

func (f *FileSystem) Write(path string, buff []byte, ofst int64) (n int, err error) {
	if err := f.OpenInWrite(path); err != nil {
		return 0, err
	}
	link, err := f.linkRepo.GetByPath(path)
	if err != nil {
		return 0, err
	}
	id := link.ObjectId
	n, err = f.objectService.Write(id, buff, ofst)
	if link, _ := f.linkRepo.GetByPath(path); link.Size < ofst+int64(len(buff)) {
		link.Size = ofst + int64(len(buff))
		err = f.linkRepo.Update(path, link)
		if err != nil {
			return 0, err
		}
	}
	return
}

func (f *FileSystem) changeFileId(path string) (newId string, err error) {
	link, err := f.linkRepo.GetByPath(path)
	if err != nil {
		return "", err
	}
	oldId := link.ObjectId
	newId, _ = core.NewUid()
	link.ObjectId = newId
	err = f.linkRepo.Update(path, link)
	if err != nil {
		return "", err
	}
	err = f.objectService.Move(oldId, newId)
	if err != nil {
		return "", err
	}
	return newId, nil
}

func (f *FileSystem) Read(path string, buff []byte, ofst int64) (n int, err error) {
	link, err := f.linkRepo.GetByPath(path)
	if err != nil {
		return 0, err
	}
	vaultLink, err := f.getVaultLink(path)
	if err != nil {
		return 0, err
	}
	keyId, err := f.objectService.GetKeyIdByObjectId(link.ObjectId)
	if err != nil {
		return 0, err
	}
	key, err := f.keyService.Get(keyId, vaultLink.VaultId)
	return f.objectService.Read(link.ObjectId, buff, ofst, key)
}

func (f *FileSystem) Resize(path string, size int64) (err error) {
	if err := f.OpenInWrite(path); err != nil {
		return err
	}
	link, err := f.linkRepo.GetByPath(path)
	if err != nil {
		return err
	}
	link.Size = size
	err = f.linkRepo.Update(path, link)
	if err != nil {
		return err
	}
	err = f.objectService.Truncate(link.ObjectId, size)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileSystem) Rename(oldPath string, newPath string) (err error) {
	if filepath.Dir(oldPath) != filepath.Dir(newPath) {
		isDir, err := f.linkRepo.IsDir(oldPath)
		if err != nil {
			return err
		}
		oldVault, err := f.linkRepo.GetVaultLinkByPath(filepath.Dir(oldPath))
		if err != nil {
			return err
		}
		newVault, err := f.linkRepo.GetVaultLinkByPath(filepath.Dir(newPath))
		if err != nil {
			return err
		}
		if isDir {
			vault, err := f.linkRepo.GetVaultLinkByPath(oldPath)
			if err != nil {
				return err
			}
			err = f.keyService.MoveVault(vault.VaultId, oldVault.VaultId, newVault.VaultId)
			if err != nil {
				return err
			}
		} else {
			obj, err := f.linkRepo.GetByPath(oldPath)
			if err != nil {
				return err
			}
			keyId, err := f.objectService.GetKeyIdByObjectId(obj.ObjectId)
			err = f.keyService.MoveKey(keyId, oldVault.VaultId, newVault.VaultId)
			if err != nil {
				return err
			}
		}
	}
	return f.linkRepo.Rename(oldPath, newPath)
}

func (f *FileSystem) Commit(path string) error {
	_, ex := f.openToWrite[path]
	if ex {
		delete(f.openToWrite, path)
		link, _ := f.linkRepo.GetByPath(path)
		vaultLink, err := f.getVaultLink(path)
		if err != nil {
			return err
		}
		keyInfo, err := f.keyService.GenerateKeyInVault(vaultLink.VaultId)
		return f.objectService.Commit(link, keyInfo)
	}
	return nil
}

func (f *FileSystem) OpenInWrite(path string) error {
	_, ex := f.openToWrite[path]
	if ex == false {
		newId, err := f.changeFileId(path)
		if err != nil {
			return err
		}
		f.openToWrite[path] = openToWrite{id: newId}
	}
	return nil
}

func (f *FileSystem) getVaultLink(path string) (core.VaultLink, error) {
	dir := filepath.Dir(path)
	vaultLink, err := f.linkRepo.GetVaultLinkByPath(dir)
	return vaultLink, err
}
