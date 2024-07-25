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

func (o *ObjectRepository) IsInRepo(id string, path string) (is bool) {
	dir := filepath.Dir(path)
	p := o.GetPath(id, dir)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false
	}
	return true
}

func (o *ObjectRepository) CreateFile(id string, dir string) (*os.File, error) {
	path := o.GetPath(id, dir)
	file, _ := os.Create(path)
	return file, nil
}

func (o *ObjectRepository) OpenObject(id string, path string) (io.ReadCloser, error) {
	dir := filepath.Dir(path)
	objectPath := o.GetPath(id, dir)
	file, _ := os.Open(objectPath)
	return file, nil
}

func (o *ObjectRepository) ChangeDir(id string, oldDir string, newDir string) error {
	oldPath := o.GetPath(id, oldDir)
	newPath := o.GetPath(id, newDir)
	return os.Rename(oldPath, newPath)
}

func (o *ObjectRepository) GetPath(id string, dir string) string {
	path := filepath.Join(o.rootPath, dir, ".meta", ".object", id)
	return path
}
