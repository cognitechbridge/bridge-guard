package filesyetem

import (
	"io/fs"
	"time"
)

type FileInfo struct {
	name  string
	size  int64
	isDir bool
}

var _ fs.FileInfo = FileInfo{}

func (f FileInfo) Name() string {
	return f.name
}

func (f FileInfo) Size() int64 {
	return f.size
}

func (f FileInfo) Mode() fs.FileMode {
	return 0
}

func (f FileInfo) ModTime() time.Time {
	return time.Time{}
}

func (f FileInfo) IsDir() bool {
	return f.isDir
}

func (f FileInfo) Sys() any {
	return nil
}

type encryptChanItem struct {
	id string
}

type uploadChanItem struct {
	path string
}
