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
	file := f.OpenFsFile(path)
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
func (f *FileSystem) OpenFsFile(path string) (fsFile *FsFile) {
	return &FsFile{
		path: path,
		file: nil,
		fs:   f,
	}
}

// GetFileId retrieves the file id by path
func (f *FileSystem) GetFileId(path string) (key string, err error) {
	file := f.OpenFsFile(path)
	key, err = file.ReadId()
	return key, err
}

func (c *FsFile) WriteSize(size int64) (err error) {
	defer c.sync()()
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, size)
	_, err = c.file.WriteAt(buf.Bytes(), FsSizeOffset)
	return
}

func (c *FsFile) ReadSize() (size int64, err error) {
	defer c.sync()()
	if err != nil {
		return 0, err
	}
	buf := make([]byte, 8)
	_, err = c.file.ReadAt(buf, FsSizeOffset)
	err = binary.Read(bytes.NewReader(buf), binary.BigEndian, &size)
	if err != nil {
		return
	}
	return
}

func (c *FsFile) WriteId(key string) (err error) {
	defer c.sync()()
	if err != nil {
		return err
	}
	_, err = c.file.WriteAt([]byte(key), FsIdOffset)
	return
}

func (c *FsFile) ReadId() (key string, err error) {
	defer c.sync()()
	if err != nil {
		return "", err
	}
	id := make([]byte, 36)
	_, _ = c.file.ReadAt(id, FsIdOffset)
	return string(id), nil
}

func (c *FsFile) sync() (fu func()) {
	_ = c.open()
	return func() {
		_ = c.close()
	}
}

func (c *FsFile) open() (err error) {
	if c.file != nil {
		return nil
	}
	p := filepath.Join(c.fs.fileSystemPath, c.path)
	c.file, err = os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0666)
	return
}

func (c *FsFile) close() (err error) {
	err = c.file.Close()
	c.file = nil
	return
}
