package filesyetem

import (
	"github.com/google/uuid"
	"io"
	"os"
	"path/filepath"
)

// FileSystem implements the FileSystem interface
type FileSystem struct {
	rootPath        string
	fileSystemPath  string
	ObjectPath      string
	ObjectCachePath string
}

// NewPersistFileSystem creates a new instance of PersistFileSystem
func NewPersistFileSystem() *FileSystem {
	fs := FileSystem{}
	fs.rootPath, _ = GetRepoCtbRoot()
	fs.fileSystemPath = filepath.Join(fs.rootPath, "filesystem")
	fs.ObjectPath = filepath.Join(fs.rootPath, "object")
	fs.ObjectCachePath = filepath.Join(fs.rootPath, "cache")
	return &fs
}

// SavePath saves a serialized key in the database
func (f *FileSystem) SavePath(key string, path string) error {
	absPath := filepath.Join(f.fileSystemPath, path)
	err := os.MkdirAll(filepath.Dir(absPath), os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.Create(absPath)
	defer file.Close()
	_, err = file.Write([]byte(key))
	if err != nil {
		return err
	}
	return nil
}

// GetPath retrieves a path id by path
func (f *FileSystem) GetPath(path string) (string, error) {
	absPath := filepath.Join(f.fileSystemPath, path)
	file, err := os.Open(absPath)
	defer file.Close()
	id, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(id), nil
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

func (f *FileSystem) GetSubNames(path string) []os.FileInfo {
	p := filepath.Join(f.fileSystemPath, path)
	file, _ := os.Open(p)
	defer file.Close()
	files, _ := file.Readdir(0)
	return files
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
	_ = f.SavePath(key.String(), path)
	return
}

func (f *FileSystem) Write(path string, buff []byte, ofst int64) (n int, err error) {
	id, err := f.GetPath(path)
	if err != nil {
		return 0, err
	}
	p := filepath.Join(f.ObjectCachePath, id)
	file, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		return 0, err
	}
	n, err = file.WriteAt(buff, ofst)
	return
}

func (f *FileSystem) Read(path string, buff []byte, ofst int64) (n int, err error) {
	id, err := f.GetPath(path)
	if err != nil {
		return 0, err
	}
	p := filepath.Join(f.ObjectCachePath, id)
	file, err := os.OpenFile(p, os.O_RDONLY, 0666)
	defer file.Close()
	if err != nil {
		return 0, err
	}
	n, err = file.ReadAt(buff, ofst)
	return
}

func GetRepoCtbRoot() (string, error) {
	root, err := os.Getwd()
	if err != nil {
		return "", err
	}
	path := filepath.Join(root, ".ctb")
	return path, nil
}
