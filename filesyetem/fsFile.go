package filesyetem

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
)

type FsFile struct {
	path string
	file *os.File
	fs   *FileSystem
}

const FsSizeOffset = 0
const FsIdOffset = 8

// CreateFsFile Create FS file
func (f *FileSystem) CreateFsFile(key string, path string, size int64) error {
	absPath := filepath.Join(f.fileSystemPath, path)
	err := os.MkdirAll(filepath.Dir(absPath), os.ModePerm)
	if err != nil {
		return err
	}
	file, err := f.OpenFsFile(path)
	defer file.Close()
	if err != nil {
		return err
	}
	err = file.WriteSize(size)
	if err != nil {
		return err
	}
	err = file.WriteId(key)
	if err != nil {
		return err
	}
	return nil
}

// OpenFsFile Open FS file
func (f *FileSystem) OpenFsFile(path string) (fsFile *FsFile, err error) {
	p := filepath.Join(f.fileSystemPath, path)
	file, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0666)
	return &FsFile{
		path: path,
		file: file,
		fs:   f,
	}, nil
}

// GetFileId retrieves the file id by path
func (f *FileSystem) GetFileId(path string) (string, error) {
	file, err := f.OpenFsFile(path)
	defer file.Close()
	key, err := file.ReadId()
	return key, err
}

func (c *FsFile) Close() (err error) {
	return c.file.Close()
}

func (c *FsFile) WriteSize(size int64) (err error) {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, size)
	_, err = c.file.WriteAt(buf.Bytes(), FsSizeOffset)
	return
}

func (c *FsFile) ReadSize() (size int64, err error) {
	buf := make([]byte, 8)
	_, err = c.file.ReadAt(buf, FsSizeOffset)
	err = binary.Read(bytes.NewReader(buf), binary.BigEndian, &size)
	if err != nil {
		return
	}
	return
}

func (c *FsFile) WriteId(key string) (err error) {
	_, err = c.file.WriteAt([]byte(key), FsIdOffset)
	return
}

func (c *FsFile) ReadId() (key string, err error) {
	id := make([]byte, 36)
	_, _ = c.file.ReadAt(id, FsIdOffset)
	return string(id), nil
}

func (c *FsFile) Resize(size int64) (err error) {
	err = c.WriteSize(size)
	if err != nil {
		return err
	}
	return
}

func (c *FsFile) ReId(id string) (err error) {
	err = c.WriteId(id)
	if err != nil {
		return err
	}
	return
}
