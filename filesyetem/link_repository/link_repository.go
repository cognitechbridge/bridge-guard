package link_repository

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
)

type LinkRepository struct {
	rootPath string
	file     *os.File
}

const FsSizeOffset = 0
const FsIdOffset = 8

func New(rootPath string) *LinkRepository {
	return &LinkRepository{
		file:     nil,
		rootPath: rootPath,
	}
}

// Create link file
func (c *LinkRepository) Create(path string, key string, size int64) error {
	absPath := filepath.Join(c.rootPath, path)
	err := os.MkdirAll(filepath.Dir(absPath), os.ModePerm)
	if err != nil {
		return err
	}
	err = c.open(path)
	if err != nil {
		return err
	}
	err = c.WriteSize(path, size)
	if err != nil {
		return err
	}
	err = c.WriteId(path, key)
	if err != nil {
		return err
	}
	return nil
}

func (c *LinkRepository) WriteSize(path string, size int64) (err error) {
	defer c.openClose(path)()
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, size)
	_, err = c.file.WriteAt(buf.Bytes(), FsSizeOffset)
	return
}

func (c *LinkRepository) ReadSize(path string) (size int64, err error) {
	defer c.openClose(path)()
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

func (c *LinkRepository) WriteId(path string, key string) (err error) {
	defer c.openClose(path)()
	if err != nil {
		return err
	}
	_, err = c.file.WriteAt([]byte(key), FsIdOffset)
	return
}

func (c *LinkRepository) ReadId(path string) (key string, err error) {
	defer c.openClose(path)()
	if err != nil {
		return "", err
	}
	id := make([]byte, 36)
	_, _ = c.file.ReadAt(id, FsIdOffset)
	return string(id), nil
}

func (c *LinkRepository) openClose(path string) (fu func()) {
	_ = c.open(path)
	return func() {
		_ = c.close()
	}
}

func (c *LinkRepository) open(path string) (err error) {
	if c.file != nil {
		return nil
	}
	p := filepath.Join(c.rootPath, path)
	c.file, err = os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0666)
	return
}

func (c *LinkRepository) close() (err error) {
	err = c.file.Close()
	c.file = nil
	return
}
