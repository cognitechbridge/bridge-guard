package filesystem_service

import (
	"io/fs"
	"time"
)

type FileInfo struct {
	name  string
	size  int64
	isDir bool
	mode  fs.FileMode
}

var _ fs.FileInfo = FileInfo{}

func (f FileInfo) Name() string {
	return f.name
}

func (f FileInfo) Size() int64 {
	return f.size
}

func (f FileInfo) Mode() fs.FileMode {
	return f.mode
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
