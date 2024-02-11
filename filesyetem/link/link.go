package link

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
)

type Link struct {
	rootPath string
	path     string
	file     *os.File
}

const FsSizeOffset = 0
const FsIdOffset = 8

func New(path string, rootPath string) *Link {
	return &Link{
		path:     path,
		file:     nil,
		rootPath: rootPath,
	}
}

// Create link file
func (c *Link) Create(key string, size int64) error {
	absPath := filepath.Join(c.rootPath, c.path)
	err := os.MkdirAll(filepath.Dir(absPath), os.ModePerm)
	if err != nil {
		return err
	}
	err = c.open()
	if err != nil {
		return err
	}
	err = c.WriteSize(size)
	if err != nil {
		return err
	}
	err = c.WriteId(key)
	if err != nil {
		return err
	}
	return nil
}

func (c *Link) WriteSize(size int64) (err error) {
	defer c.sync()()
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, size)
	_, err = c.file.WriteAt(buf.Bytes(), FsSizeOffset)
	return
}

func (c *Link) ReadSize() (size int64, err error) {
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

func (c *Link) WriteId(key string) (err error) {
	defer c.sync()()
	if err != nil {
		return err
	}
	_, err = c.file.WriteAt([]byte(key), FsIdOffset)
	return
}

func (c *Link) ReadId() (key string, err error) {
	defer c.sync()()
	if err != nil {
		return "", err
	}
	id := make([]byte, 36)
	_, _ = c.file.ReadAt(id, FsIdOffset)
	return string(id), nil
}

func (c *Link) sync() (fu func()) {
	_ = c.open()
	return func() {
		_ = c.close()
	}
}

func (c *Link) open() (err error) {
	if c.file != nil {
		return nil
	}
	p := filepath.Join(c.rootPath, c.path)
	c.file, err = os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0666)
	return
}

func (c *Link) close() (err error) {
	err = c.file.Close()
	c.file = nil
	return
}
