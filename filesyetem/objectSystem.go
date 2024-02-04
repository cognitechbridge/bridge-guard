package filesyetem

import (
	"os"
	"path/filepath"
)

type ObjectSystem struct {
	downloader      Downloader
	ObjectCachePath string
	ObjectWritePath string
}

func NewObjectSystem(path string) ObjectSystem {
	return ObjectSystem{
		ObjectCachePath: path,
		ObjectWritePath: filepath.Join(path, "write"),
	}
}

func (o *ObjectSystem) moveObject(oldId string, newId string) (err error) {
	//Move to write cache path
	oldPath := filepath.Join(o.ObjectCachePath, oldId)
	newPath := filepath.Join(o.ObjectWritePath, newId)
	err = os.Rename(oldPath, newPath)
	if err != nil {
		return
	}
	//Create link
	err = o.createWriteLink(newId)
	if err != nil {
		return
	}
	return nil
}

func (o *ObjectSystem) writeObject(id string, buff []byte, ofst int64) (n int, err error) {
	p := filepath.Join(o.ObjectWritePath, id)
	if _, err := os.Stat(p); os.IsNotExist(err) { // If file not exist
		err = o.downloader.Download(id)
		if err != nil {
			return 0, err
		}
	}
	file, err := os.OpenFile(p, os.O_RDWR, 0666)
	defer file.Close()
	if err != nil {
		return 0, err
	}
	n, err = file.WriteAt(buff, ofst)
	return
}

func (o *ObjectSystem) truncateObject(id string, size int64) (err error) {
	p := filepath.Join(o.ObjectWritePath, id)
	file, err := os.OpenFile(p, os.O_RDWR, 0666)
	err = file.Truncate(size)
	if err != nil {
		return err
	}
	return nil
}

func (o *ObjectSystem) createObject(id string) (err error) {
	objWritePath := filepath.Join(o.ObjectWritePath, id)
	objFile, err := os.Create(objWritePath)
	objFile.Close()
	if err != nil {
		return
	}
	err = o.createWriteLink(id)
	if err != nil {
		return
	}
	return nil
}

func (o *ObjectSystem) readObject(id string, buff []byte, ofst int64) (n int, err error) {
	p := filepath.Join(o.ObjectCachePath, id)

	if _, err := os.Stat(p); os.IsNotExist(err) {
		err = o.downloader.Download(id)
		if err != nil {
			return 0, err
		}
	}

	file, err := os.OpenFile(p, os.O_RDONLY, 0666)
	defer file.Close()
	if err != nil {
		return 0, err
	}
	n, err = file.ReadAt(buff, ofst)
	return
}

func (o *ObjectSystem) createWriteLink(id string) (err error) {
	objWritePath := filepath.Join(o.ObjectWritePath, id)
	objFilePath := filepath.Join(o.ObjectCachePath, id)
	err = os.Link(objWritePath, objFilePath)
	return err
}
