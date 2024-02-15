package repositories

import (
	"io"
	"os"
	"path/filepath"
)

type ObjectRepository struct {
	rootPath string
}

func NewObjectRepository(rootPath string) ObjectRepository {
	return ObjectRepository{
		rootPath: rootPath,
	}
}

func (o *ObjectRepository) IsInRepo(id string) (is bool) {
	p := filepath.Join(o.rootPath, id)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false
	}
	return true
}

func (o *ObjectRepository) CreateFile(id string) (*os.File, error) {
	path := filepath.Join(o.rootPath, id)
	file, _ := os.Create(path)
	return file, nil
}

func (o *ObjectRepository) OpenObject(id string) (io.ReadCloser, error) {
	path := filepath.Join(o.rootPath, id)
	file, _ := os.Open(path)
	return file, nil
}

func (o *ObjectRepository) GetPath(id string) (string, error) {
	path := filepath.Join(o.rootPath, id)
	return path, nil
}
