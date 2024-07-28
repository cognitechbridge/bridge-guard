package repositories

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var (
	ErrRemoveFileFromCache = fmt.Errorf("error removing file from cache")
)

type ObjectCacheRepository struct {
	resolver       func(id string, writer io.Writer) (err error)
	readPath       string
	writePath      string
	committingList map[string]struct{}
}

func NewObjectCacheRepository(path string) ObjectCacheRepository {
	writePath := filepath.Join(path, "Write")
	err := os.MkdirAll(writePath, os.ModePerm)
	if err != nil {
		panic(err)
	}
	return ObjectCacheRepository{
		readPath:       path,
		writePath:      writePath,
		committingList: make(map[string]struct{}),
	}
}

// @TODO: Refactor this
func (o *ObjectCacheRepository) Move(oldId string, newId string) (err error) {
	//Move to Write cache path
	oldPath := filepath.Join(o.readPath, oldId)
	newPath := filepath.Join(o.writePath, newId)
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

func (o *ObjectCacheRepository) CacheObjectWriter(id string) (io.WriteCloser, error) {
	p := filepath.Join(o.readPath, id)
	file, err := os.Create(p)
	return file, err
}

func (o *ObjectCacheRepository) Write(id string, buff []byte, ofst int64) (n int, err error) {
	p := filepath.Join(o.writePath, id)
	file, err := os.OpenFile(p, os.O_RDWR, 0666)
	if err != nil {
		return 0, fmt.Errorf("file is not in write cache: %v", err)
	}
	defer file.Close()
	if err != nil {
		return 0, err
	}
	n, err = file.WriteAt(buff, ofst)
	return
}

func (o *ObjectCacheRepository) Truncate(id string, size int64) (err error) {
	p := filepath.Join(o.writePath, id)
	file, err := os.OpenFile(p, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	err = file.Truncate(size)
	if err != nil {
		return err
	}
	return nil
}

func (o *ObjectCacheRepository) Create(id string) (err error) {
	objWritePath := filepath.Join(o.writePath, id)
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

func (o *ObjectCacheRepository) Read(id string, buff []byte, ofst int64) (n int, err error) {
	p := filepath.Join(o.readPath, id)

	if _, err := os.Stat(p); os.IsNotExist(err) {
		err = o.resolverFile(id)
		if err != nil {
			return 0, err
		}
	}

	file, err := os.OpenFile(p, os.O_RDONLY, 0666)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	n, err = file.ReadAt(buff, ofst)
	return
}

func (o *ObjectCacheRepository) IsInCache(id string) (is bool) {
	p := filepath.Join(o.readPath, id)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false
	}
	return true
}

func (o *ObjectCacheRepository) Flush(id string) (err error) {
	delete(o.committingList, id)
	p := filepath.Join(o.writePath, id)
	err = os.Remove(p)
	if err != nil {
		return
	}
	return
}

func (o *ObjectCacheRepository) AsFile(id string) (file *os.File, err error) {
	p := filepath.Join(o.readPath, id)
	file, err = os.OpenFile(p, os.O_RDONLY, 0666)
	return file, err
}

func (o *ObjectCacheRepository) createWriteLink(id string) (err error) {
	objWritePath := filepath.Join(o.writePath, id)
	objFilePath := filepath.Join(o.readPath, id)
	err = os.Link(objWritePath, objFilePath)
	return err
}

func (o *ObjectCacheRepository) resolverFile(id string) (err error) {
	file, _ := os.Create(filepath.Join(o.readPath, id))
	defer file.Close()
	err = o.resolver(id, file)
	return
}

// RemoveFromCache removes the object with the specified ID from the cache.
// It returns an error if the removal operation fails.
// If the object is not in the cache, it returns nil (no error).
func (o *ObjectCacheRepository) RemoveFromCache(id string) error {
	// If the object is in the list of committed objects, we should wait for it to be committed.
	_, committed := o.committingList[id]
	if committed {
		return nil
	}

	p := filepath.Join(o.readPath, id)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return nil
	}
	err := os.Remove(p)
	if err != nil {
		return ErrRemoveFileFromCache
	}
	return nil
}

// IsOpenForWrite returns true if the object is in the write cache.
func (o *ObjectCacheRepository) IsOpenForWrite(id string) bool {
	p := filepath.Join(o.writePath, id)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false
	}
	_, committed := o.committingList[id]
	return !committed
}

// AdToCommitting marks the object as committed in the write cache.
func (o *ObjectCacheRepository) AdToCommitting(id string) {
	o.committingList[id] = struct{}{}
}
