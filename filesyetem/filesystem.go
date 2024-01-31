package filesyetem

import (
	"github.com/google/uuid"
	"os"
	"path/filepath"
)

// FileSystem implements the FileSystem interface
type FileSystem struct {
	UploadQueue *UploadQueue
	downloader  Downloader

	rootPath        string
	fileSystemPath  string
	ObjectPath      string
	ObjectCachePath string
}

type Downloader interface {
	Download(id string) error
}

// NewPersistFileSystem creates a new instance of PersistFileSystem
func NewPersistFileSystem(dn Downloader) *FileSystem {
	fs := FileSystem{}
	fs.rootPath, _ = GetRepoCtbRoot()
	fs.fileSystemPath = filepath.Join(fs.rootPath, "filesystem")
	fs.ObjectPath = filepath.Join(fs.rootPath, "object")
	fs.ObjectCachePath = filepath.Join(fs.rootPath, "cache")
	fs.UploadQueue = NewUploadQueue()
	fs.downloader = dn
	return &fs
}

func (f *FileSystem) CreateDir(path string) error {
	absPath := filepath.Join(f.fileSystemPath, path)
	err := os.MkdirAll(absPath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileSystem) RemovePath(path string) error {
	absPath := filepath.Join(f.fileSystemPath, path)
	err := os.Remove(absPath)
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

type FileInfo struct {
	Name  string
	Size  int64
	IsDir bool
}

func (f *FileSystem) GetSubFiles(path string) []FileInfo {
	p := filepath.Join(f.fileSystemPath, path)
	file, _ := os.Open(p)
	defer file.Close()
	subFiles, _ := file.Readdir(0)
	infos := make([]FileInfo, 0)
	for _, subFile := range subFiles {
		if subFile.IsDir() {
			infos = append(infos, FileInfo{
				IsDir: true,
				Name:  subFile.Name(),
			})
			continue
		} else {
			p := filepath.Join(path, subFile.Name())
			file := f.OpenFsFile(p)
			size, _ := file.ReadSize()
			infos = append(infos, FileInfo{
				IsDir: false,
				Name:  subFile.Name(),
				Size:  size,
			})
		}

	}
	return infos
}

func (f *FileSystem) RemoveDir(path string) {
	p := filepath.Join(f.fileSystemPath, path)
	os.Remove(p)
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
	f.UploadQueue.Enqueue(path)
	return
}

func (f *FileSystem) Write(path string, buff []byte, ofst int64) (n int, err error) {
	if !f.UploadQueue.IsInQueue(path) {
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
	f.UploadQueue.Enqueue(path)
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

func (f *FileSystem) Rename(oldPath string, newPath string) error {
	o := filepath.Join(f.fileSystemPath, oldPath)
	n := filepath.Join(f.fileSystemPath, newPath)
	err := os.Rename(o, n)
	if err != nil {
		return err
	}
	f.UploadQueue.Rename(oldPath, newPath)
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
