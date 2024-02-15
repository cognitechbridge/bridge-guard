package filesyetem_service

import (
	"ctb-cli/repositories"
	"ctb-cli/services/object_service"
	"fmt"
	"github.com/google/uuid"
	"io/fs"
	"os"
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
	rootPath      string
	linkRepoPath  string
}

var _ FileSystemHandler = &FileSystem{}

// NewFileSystem creates a new instance of PersistFileSystem
func NewFileSystem(objectSerivce object_service.Service) *FileSystem {
	fileSys := FileSystem{
		objectService: objectSerivce,
	}

	fileSys.rootPath, _ = GetRepoCtbRoot()
	fileSys.linkRepoPath = filepath.Join(fileSys.rootPath, "filesystem")

	fileSys.linkRepo = repositories.NewLinkRepository(fileSys.linkRepoPath)

	return &fileSys
}

func (f *FileSystem) CreateDir(path string) (err error) {
	absPath := filepath.Join(f.linkRepoPath, path)
	err = os.MkdirAll(absPath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileSystem) RemovePath(path string) (err error) {
	absPath := filepath.Join(f.linkRepoPath, path)
	err = os.Remove(absPath)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileSystem) PathExist(path string) (bool, error) {
	absPath := filepath.Join(f.linkRepoPath, path)
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return false, nil
	} else {
		return true, nil
	}
}

func (f *FileSystem) IsDir(path string) bool {
	p := filepath.Join(f.linkRepoPath, path)
	fileInfo, _ := os.Stat(p)
	return fileInfo.IsDir()
}

func (f *FileSystem) GetSubFiles(path string) (res []fs.FileInfo, err error) {
	p := filepath.Join(f.linkRepoPath, path)
	file, err := os.Open(p)
	defer file.Close()
	if err != nil {
		return nil, fmt.Errorf("error opening dir to Read sub files: %v", err)
	}
	subFiles, _ := file.Readdir(0)
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
			size, err := f.linkRepo.ReadSize(p)
			if err != nil {
				return nil, fmt.Errorf("error reading file size: %v", err)
			}
			var info fs.FileInfo = FileInfo{
				isDir: false,
				name:  subFile.Name(),
				size:  size,
			}
			infos = append(infos, info)
		}

	}
	return infos, nil
}

func (f *FileSystem) RemoveDir(path string) (err error) {
	p := filepath.Join(f.linkRepoPath, path)
	err = os.Remove(p)
	return err
}

func (f *FileSystem) CreateFile(path string) (err error) {
	key, err := uuid.NewV7()
	if err != nil {
		return err
	}
	_ = f.linkRepo.Create(path, key.String(), 0)
	err = f.objectService.Create(key.String())
	if err != nil {
		return err
	}
	return
}

func (f *FileSystem) Write(path string, buff []byte, ofst int64) (n int, err error) {
	id, err := f.linkRepo.ReadId(path)
	if !f.objectService.IsInQueue(id) {
		id, err = f.changeFileId(path)
		if err != nil {
			return 0, err
		}
	}
	n, err = f.objectService.Write(id, buff, ofst)
	if size, _ := f.linkRepo.ReadSize(path); size < ofst+int64(len(buff)) {
		err = f.linkRepo.WriteSize(path, ofst+int64(len(buff)))
		if err != nil {
			return 0, err
		}
	}
	return
}

func (f *FileSystem) changeFileId(path string) (newId string, err error) {
	oldId, err := f.linkRepo.ReadId(path)
	if err != nil {
		return "", err
	}
	uui, _ := uuid.NewV7()
	newId = uui.String()
	err = f.linkRepo.WriteId(path, newId)
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
	id, err := f.linkRepo.ReadId(path)
	if err != nil {
		return 0, err
	}
	return f.objectService.Read(id, buff, ofst)
}

func (f *FileSystem) Resize(path string, size int64) (err error) {
	err = f.linkRepo.WriteSize(path, size)
	if err != nil {
		return err
	}
	id, err := f.linkRepo.ReadId(path)
	if err != nil {
		return err
	}
	err = f.objectService.Truncate(id, size)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileSystem) Rename(oldPath string, newPath string) (err error) {
	o := filepath.Join(f.linkRepoPath, oldPath)
	n := filepath.Join(f.linkRepoPath, newPath)
	err = os.Rename(o, n)
	if err != nil {
		return err
	}
	return nil
}

func GetRepoCtbRoot() (string, error) {
	root, err := os.Getwd()
	if err != nil {
		return "", err
	}
	path := filepath.Join(root, ".ctb")
	return path, nil
}
