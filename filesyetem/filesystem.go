package filesyetem

import (
	"ctb-cli/config"
	"io"
	"os"
	"path/filepath"
)

// FileSystem implements the FileSystem interface
type FileSystem struct {
}

// NewPersistFileSystem creates a new instance of PersistFileSystem
func NewPersistFileSystem() *FileSystem {
	return &FileSystem{}
}

// SavePath saves a serialized key in the database
func (f *FileSystem) SavePath(key string, path string) error {
	absPath, err := getAbsPath(path)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(absPath), os.ModePerm)
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
	absPath, err := getAbsPath(path)
	if err != nil {
		return "", err
	}
	file, err := os.Open(absPath)
	defer file.Close()
	id, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(id), nil
}

func (f *FileSystem) RemovePath(path string) error {
	absPath, err := getAbsPath(path)
	if err != nil {
		return err
	}
	err = os.Remove(absPath)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileSystem) PathExist(path string) (bool, error) {
	absPath, err := getAbsPath(path)
	if err != nil {
		return false, err
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return false, nil
	} else {
		return true, nil
	}
}

func getAbsPath(path string) (string, error) {
	basePath, err := getFilesysPath()
	if err != nil {
		return "", err
	}
	absPath := filepath.Join(basePath, path)
	return absPath, nil
}

func getFilesysPath() (string, error) {
	rootPath, err := config.GetRepoCtbRoot()
	if err != nil {
		return "", err
	}
	path := filepath.Join(rootPath, "filesystem")
	return path, nil
}
