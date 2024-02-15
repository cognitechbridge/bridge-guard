package filesyetem

import (
	"ctb-cli/filesyetem/link_repository"
	"ctb-cli/filesyetem/object"
	"ctb-cli/filesyetem/object_cache"
	"fmt"
	"github.com/google/uuid"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// FileSystem implements the FileSystem interface
type FileSystem struct {
	//interfaces
	objectService object.Service
	downloader    Downloader

	//internal queues and channels
	encryptChan  chan encryptChanItem
	uploadChan   chan uploadChanItem
	encryptQueue *EncryptQueue

	objectCacheSystem object_cache.ObjectCache
	linkRepo          *link_repository.LinkRepository

	//path
	rootPath       string
	fileSystemPath string
	ObjectPath     string
}

type Downloader interface {
	Download(id string, writeAt io.WriterAt) error
	Upload(reader io.Reader, fileId string) error
}

// NewFileSystem creates a new instance of PersistFileSystem
func NewFileSystem(dn Downloader, objectSerivce object.Service) *FileSystem {
	fileSys := FileSystem{
		downloader:    dn,
		objectService: objectSerivce,

		encryptChan: make(chan encryptChanItem, 10),
		uploadChan:  make(chan uploadChanItem, 10),
	}

	fileSys.rootPath, _ = GetRepoCtbRoot()
	fileSys.fileSystemPath = filepath.Join(fileSys.rootPath, "filesystem")

	fileSys.encryptQueue = fileSys.NewEncryptQueue()
	fileSys.objectCacheSystem = object_cache.New(
		filepath.Join(fileSys.rootPath, "cache"),
	)

	fileSys.linkRepo = link_repository.New(fileSys.fileSystemPath)

	go fileSys.StartEncryptRoutine()
	go fileSys.StartUploadRoutine()

	return &fileSys
}

func (f *FileSystem) CreateDir(path string) (err error) {
	absPath := filepath.Join(f.fileSystemPath, path)
	err = os.MkdirAll(absPath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileSystem) RemovePath(path string) (err error) {
	absPath := filepath.Join(f.fileSystemPath, path)
	err = os.Remove(absPath)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileSystem) PathExist(path string) (bool, error) {
	absPath := filepath.Join(f.fileSystemPath, path)
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return false, nil
	} else {
		return true, nil
	}
}

func (f *FileSystem) IsDir(path string) bool {
	p := filepath.Join(f.fileSystemPath, path)
	fileInfo, _ := os.Stat(p)
	return fileInfo.IsDir()
}

func (f *FileSystem) GetSubFiles(path string) (res []fs.FileInfo, err error) {
	p := filepath.Join(f.fileSystemPath, path)
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
	p := filepath.Join(f.fileSystemPath, path)
	err = os.Remove(p)
	return err
}

func (f *FileSystem) CreateFile(path string) (err error) {
	key, err := uuid.NewV7()
	if err != nil {
		return
	}
	_ = f.linkRepo.Create(path, key.String(), 0)
	err = f.objectCacheSystem.Create(key.String())
	if err != nil {
		return
	}
	f.encryptQueue.Enqueue(path)
	return
}

func (f *FileSystem) Write(path string, buff []byte, ofst int64) (n int, err error) {
	id, err := f.linkRepo.ReadId(path)
	if !f.encryptQueue.IsInQueue(path) {
		id, err = f.changeFileId(path)
		if err != nil {
			return 0, err
		}
	}
	n, err = f.objectCacheSystem.Write(id, buff, ofst)
	if size, _ := f.linkRepo.ReadSize(path); size < ofst+int64(len(buff)) {
		err = f.linkRepo.WriteSize(path, ofst+int64(len(buff)))
		if err != nil {
			return 0, err
		}
	}
	f.encryptQueue.Enqueue(path)
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
	err = f.objectCacheSystem.Move(oldId, newId)
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
	err = f.objectCacheSystem.Truncate(id, size)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileSystem) Rename(oldPath string, newPath string) (err error) {
	o := filepath.Join(f.fileSystemPath, oldPath)
	n := filepath.Join(f.fileSystemPath, newPath)
	err = os.Rename(o, n)
	if err != nil {
		return err
	}
	f.encryptQueue.Rename(oldPath, newPath)
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
