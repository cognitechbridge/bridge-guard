package repositories

import (
	"ctb-cli/core"
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

func (o *ObjectRepository) IsInRepo(link core.Link) (is bool) {
	p := o.GetPath(link.Id(), link.Path)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false
	}
	return true
}

func (o *ObjectRepository) CreateFile(link core.Link) (*os.File, error) {
	objectPath := o.GetPath(link.Id(), link.Path)
	file, _ := os.Create(objectPath)
	return file, nil
}

func (o *ObjectRepository) OpenObject(link core.Link) (io.ReadCloser, error) {
	objectPath := o.GetPath(link.Id(), link.Path)
	file, _ := os.Open(objectPath)
	return file, nil
}

func (o *ObjectRepository) ChangePath(link core.Link, newPath string) error {
	if o.GetPath(link.Id(), link.Path) != o.GetPath(link.Id(), newPath) {
		oldObjectPath := o.GetPath(link.Id(), link.Path)
		newObjectPath := o.GetPath(link.Id(), newPath)
		return os.Rename(oldObjectPath, newObjectPath)
	}
	return nil
}

func (o *ObjectRepository) GetPath(id string, path string) string {
	dir := filepath.Dir(path)
	res := filepath.Join(o.rootPath, dir, ".meta", ".object", id)
	return res
}
