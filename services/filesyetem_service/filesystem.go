package filesyetem_service

import (
	"ctb-cli/repositories"
	"ctb-cli/services/object_service"
	"ctb-cli/core"
	"fmt"
	"io/fs"
	"path/filepath"
)

type FileSystemHandler interface {
	GetSubFiles(path string) (res []fs.FileInfo, err error)
	CreateFile(path string) (err error)
	CreateDir(path string) (err error)
	RemoveDir(path string) (err error)
	Write(path string, buff []byte, ofst int64) (n int, err error)
	Read(path string, buff []byte, ofst int64) (n int, err error)
	Rename(oldPath string, newPath string) (err error)
	RemovePath(path string) (err error)
	Resize(path string, size int64) (err error)
}

// FileSystem implements the FileSystem interface
type FileSystem struct {
	objectService object_service.Service
	linkRepo      *repositories.LinkRepository
}

var _ FileSystemHandler = &FileSystem{}

// NewFileSystem creates a new instance of PersistFileSystem
func NewFileSystem(objectSerivce object_service.Service, linkRepository *repositories.LinkRepository) *FileSystem {
	fileSys := FileSystem{
		objectService: objectSerivce,
		linkRepo:      linkRepository,
	}

	return &fileSys
}

func (f *FileSystem) CreateDir(path string) (err error) {
	return f.linkRepo.CreateDir(path)
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
	key, err := core.NewUid()
	if err != nil {
		return err
	}
	_ = f.linkRepo.Create(path, core.Link{
		ObjectId: key,
		Size:     0,
	})
	err = f.objectService.Create(key)
	if err != nil {
		return err
	}
	return
}

func (f *FileSystem) Write(path string, buff []byte, ofst int64) (n int, err error) {
	link, err := f.linkRepo.GetByPath(path)
	if err != nil {
		return 0, err
	}
	id := link.ObjectId
	if !f.objectService.IsInQueue(id) {
		id, err = f.changeFileId(path)
		if err != nil {
			return 0, err
		}
	}
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
	return f.objectService.Read(link.ObjectId, buff, ofst)
}

func (f *FileSystem) Resize(path string, size int64) (err error) {
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
	return f.linkRepo.Rename(oldPath, newPath)
}
