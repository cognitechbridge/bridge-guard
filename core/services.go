package core

import "io/fs"

type ObjectService interface {
	Read(id string, buff []byte, ofst int64) (int, error)
	Write(id string, buff []byte, ofst int64) (int, error)
	Create(id string) error
	Move(oldId string, newId string) error
	Truncate(id string, size int64) error
	IsInQueue(id string) bool
	GetKeyIdByObjectId(id string) (string, error)
}

type FileSystemService interface {
	GetSubFiles(path string) (res []fs.FileInfo, err error)
	CreateFile(path string) (err error)
	CreateDir(path string) (err error)
	RemoveDir(path string) (err error)
	Write(path string, buff []byte, ofst int64) (n int, err error)
	Read(path string, buff []byte, ofst int64) (n int, err error)
	Rename(oldPath string, newPath string) (err error)
	RemovePath(path string) (err error)
	Resize(path string, size int64) (err error)
}
