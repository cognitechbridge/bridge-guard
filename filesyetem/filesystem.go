package filesyetem

import (
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
	encryptor  Encryptor
	downloader Downloader

	//out channels
	UploadChan chan string

	//internal queues and channels
	encryptChan  chan string
	encryptQueue *EncryptQueue

	//path
	rootPath        string
	fileSystemPath  string
	ObjectPath      string
	ObjectCachePath string
}

type Downloader interface {
	Download(id string) error
}

type Encryptor interface {
	Encrypt(reader io.Reader, fileId string) (read io.Reader, err error)
}

// NewFileSystem creates a new instance of PersistFileSystem
func NewFileSystem(dn Downloader, en Encryptor) *FileSystem {
	fileSys := FileSystem{
		downloader:   dn,
		encryptor:    en,
		encryptQueue: NewEncryptQueue(),

		encryptChan: make(chan string, 10),
		UploadChan:  make(chan string, 10),
	}

	fileSys.encryptChan = make(chan string, 100)

	fileSys.rootPath, _ = GetRepoCtbRoot()
	fileSys.fileSystemPath = filepath.Join(fileSys.rootPath, "filesystem")
	fileSys.ObjectPath = filepath.Join(fileSys.rootPath, "object")
	fileSys.ObjectCachePath = filepath.Join(fileSys.rootPath, "cache")

	go fileSys.encryptQueue.StartQueueRoutine(fileSys.encryptChan)
	go fileSys.StartEncryptRoutine()

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
		return nil, fmt.Errorf("error opening dir to read sub files: %v", err)
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
			file := f.OpenFsFile(p)
			size, err := file.ReadSize()
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
	_ = f.CreateFsFile(key.String(), path, 0)
	objPath := filepath.Join(f.ObjectCachePath, key.String())
	objFile, err := os.Create(objPath)
	objFile.Close()
	if err != nil {
		return
	}
	f.encryptQueue.Enqueue(path)
	return
}

func (f *FileSystem) Write(path string, buff []byte, ofst int64) (n int, err error) {
	if !f.encryptQueue.IsInQueue(path) {
		i, err := f.changeFileId(path)
		if err != nil {
			return i, err
		}
	}
	return f.writeToCache(path, buff, ofst)
}

func (f *FileSystem) changeFileId(path string) (int, error) {
	fsFile := f.OpenFsFile(path)
	oldId, err := fsFile.ReadId()
	if err != nil {
		return 0, err
	}
	uui, _ := uuid.NewV7()
	newId := uui.String()
	err = fsFile.WriteId(newId)
	if err != nil {
		return 0, err
	}
	oldPath := filepath.Join(f.ObjectCachePath, oldId)
	newPath := filepath.Join(f.ObjectCachePath, newId)
	err = os.Rename(oldPath, newPath)
	if err != nil {
		return 0, err
	}
	return 0, nil
}

func (f *FileSystem) writeToCache(path string, buff []byte, ofst int64) (n int, err error) {
	fsFile := f.OpenFsFile(path)
	id, _ := fsFile.ReadId()
	p := filepath.Join(f.ObjectCachePath, id)
	file, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		return 0, err
	}
	stat, _ := file.Stat()
	if stat.Size() < ofst+int64(len(buff)) {
		err := fsFile.WriteSize(ofst + int64(len(buff)))
		if err != nil {
			return 0, err
		}
	}
	n, err = file.WriteAt(buff, ofst)
	f.encryptQueue.Enqueue(path)
	return
}

func (f *FileSystem) Read(path string, buff []byte, ofst int64) (n int, err error) {
	id, err := f.GetFileId(path)
	if err != nil {
		return 0, err
	}
	p := filepath.Join(f.ObjectCachePath, id)

	if _, err := os.Stat(p); os.IsNotExist(err) {
		f.downloader.Download(id)
	}

	file, err := os.OpenFile(p, os.O_RDONLY, 0666)
	defer file.Close()
	if err != nil {
		return 0, err
	}
	n, err = file.ReadAt(buff, ofst)
	return
}

func (f *FileSystem) Resize(path string, size int64) (err error) {
	fs := f.OpenFsFile(path)
	err = fs.WriteSize(size)
	if err != nil {
		return err
	}

	id, err := fs.ReadId()
	p := filepath.Join(f.ObjectCachePath, id)
	file, err := os.OpenFile(p, os.O_RDWR, 0666)
	err = file.Truncate(size)
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
