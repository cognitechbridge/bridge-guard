package filesyetem

import (
	"ctb-cli/config"
	"os"
	"path/filepath"
)

// FileSystem implements the FileSystem interface
type FileSystem struct {
	rootPath string
}

//type FileInfo struct {
//	name  string
//	isDir bool
//}

// NewPersistFileSystem creates a new instance of PersistFileSystem
func NewPersistFileSystem() *FileSystem {
	fs := FileSystem{}
	fs.rootPath, _ = getFilesysPath()
	return &fs
}

// SavePath saves a serialized key in the database
func (f *FileSystem) SavePath(key string, path string) error {
	absPath := filepath.Join(f.rootPath, path)
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

func (f *FileSystem) CreateDir(path string) error {
	absPath := filepath.Join(f.rootPath, path)
	err := os.MkdirAll(absPath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileSystem) RemovePath(path string) error {
	absPath := filepath.Join(f.rootPath, path)
	err := os.Remove(absPath)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileSystem) PathExist(path string) (bool, error) {
	absPath := filepath.Join(f.rootPath, path)
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return false, nil
	} else {
		return true, nil
	}
}

func (f *FileSystem) IsDir(path string) bool {
	p := filepath.Join(f.rootPath, path)
	fileInfo, _ := os.Stat(p)
	return fileInfo.IsDir()
}

func (f *FileSystem) GetSubNames(path string) []os.FileInfo {
	p := filepath.Join(f.rootPath, path)
	file, _ := os.Open(p)
	defer file.Close()
	files, _ := file.Readdir(0)
	return files
}

func (f *FileSystem) RemoveDir(path string) {
	p := filepath.Join(f.rootPath, path)
	os.Remove(p)
}

func getFilesysPath() (string, error) {
	rootPath, err := config.GetRepoCtbRoot()
	if err != nil {
		return "", err
	}
	path := filepath.Join(rootPath, "filesystem")
	return path, nil
}
