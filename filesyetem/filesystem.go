package filesyetem

import (
	"ctb-cli/filesyetem/link"
	"ctb-cli/filesyetem/object"
	"fmt"
	"github.com/google/uuid"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// FileSystem implements the FileSystem interface
type FileSystem struct {
	//interfaces
	fileCrypto FileCrypto
	downloader Downloader

	//internal queues and channels
	encryptChan  chan encryptChanItem
	uploadChan   chan uploadChanItem
	encryptQueue *EncryptQueue

	objectCacheSystem object.Object

	//path
	rootPath       string
	fileSystemPath string
	ObjectPath     string
}

type Downloader interface {
	Download(id string, writeAt io.WriterAt) error
	Upload(reader io.Reader, fileId string) error
}

type FileCrypto interface {
	Encrypt(writer io.Writer, fileId string) (write io.WriteCloser, err error)
	Decrypt(reader io.Reader, fileId string) (read io.Reader, err error)
}

// NewFileSystem creates a new instance of PersistFileSystem
func NewFileSystem(dn Downloader, fileCrypto FileCrypto) *FileSystem {
	fileSys := FileSystem{
		downloader: dn,
		fileCrypto: fileCrypto,

		encryptChan: make(chan encryptChanItem, 10),
		uploadChan:  make(chan uploadChanItem, 10),
	}

	fileSys.rootPath, _ = GetRepoCtbRoot()
	fileSys.fileSystemPath = filepath.Join(fileSys.rootPath, "filesystem")
	fileSys.ObjectPath = filepath.Join(fileSys.rootPath, "object")

	fileSys.encryptQueue = fileSys.NewEncryptQueue()
	fileSys.objectCacheSystem = object.New(
		filepath.Join(fileSys.rootPath, "cache"),
		fileSys.ObjectResolver,
	)

	go fileSys.StartEncryptRoutine()
	go fileSys.StartUploadRoutine()

	return &fileSys
}

func (f *FileSystem) CreateDir(path string) (err error) {
	absPath := filepath.Join(f.fileSystemPath, path)
	err = os.MkdirAll(absPath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileSystem) RemovePath(path string) (err error) {
	absPath := filepath.Join(f.fileSystemPath, path)
	err = os.Remove(absPath)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileSystem) PathExist(path string) (bool, error) {
	absPath := filepath.Join(f.fileSystemPath, path)
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return false, nil
	} else {
		return true, nil
	}
}

func (f *FileSystem) IsDir(path string) bool {
	p := filepath.Join(f.fileSystemPath, path)
	fileInfo, _ := os.Stat(p)
	return fileInfo.IsDir()
}

func (f *FileSystem) GetSubFiles(path string) (res []fs.FileInfo, err error) {
	p := filepath.Join(f.fileSystemPath, path)
	file, err := os.Open(p)
	defer file.Close()
	if err != nil {
		return nil, fmt.Errorf("error opening dir to Read sub files: %v", err)
	}
	subFiles, _ := file.Readdir(0)
	var infos []fs.FileInfo
	for _, subFile := range subFiles {
		if subFile.IsDir() {
			var info fs.FileInfo = FileInfo{
				isDir: true,
				name:  subFile.Name(),
			}
			infos = append(infos, info)
			continue
		} else {
			p := filepath.Join(path, subFile.Name())
			file := f.openLinkFile(p)
			size, err := file.ReadSize()
			if err != nil {
				return nil, fmt.Errorf("error reading file size: %v", err)
			}
			var info fs.FileInfo = FileInfo{
				isDir: false,
				name:  subFile.Name(),
				size:  size,
			}
			infos = append(infos, info)
		}

	}
	return infos, nil
}

func (f *FileSystem) RemoveDir(path string) (err error) {
	p := filepath.Join(f.fileSystemPath, path)
	err = os.Remove(p)
	return err
}

func (f *FileSystem) CreateFile(path string) (err error) {
	key, err := uuid.NewV7()
	if err != nil {
		return
	}
	_ = f.openLinkFile(path).Create(key.String(), 0)
	err = f.objectCacheSystem.Create(key.String())
	if err != nil {
		return
	}
	f.encryptQueue.Enqueue(path)
	return
}

func (f *FileSystem) Write(path string, buff []byte, ofst int64) (n int, err error) {
	fsFile := f.openLinkFile(path)
	id, err := fsFile.ReadId()
	if !f.encryptQueue.IsInQueue(path) {
		id, err = f.changeFileId(path)
		if err != nil {
			return 0, err
		}
	}
	n, err = f.objectCacheSystem.Write(id, buff, ofst)
	if size, _ := fsFile.ReadSize(); size < ofst+int64(len(buff)) {
		err = fsFile.WriteSize(ofst + int64(len(buff)))
		if err != nil {
			return 0, err
		}
	}
	f.encryptQueue.Enqueue(path)
	return
}

func (f *FileSystem) changeFileId(path string) (newId string, err error) {
	fsFile := f.openLinkFile(path)
	oldId, err := fsFile.ReadId()
	if err != nil {
		return "", err
	}
	uui, _ := uuid.NewV7()
	newId = uui.String()
	err = fsFile.WriteId(newId)
	if err != nil {
		return "", err
	}
	err = f.objectCacheSystem.Move(oldId, newId)
	if err != nil {
		return "", err
	}
	return newId, nil
}

func (f *FileSystem) Read(path string, buff []byte, ofst int64) (n int, err error) {
	id, err := f.openLinkFile(path).ReadId()
	if err != nil {
		return 0, err
	}
	return f.objectCacheSystem.Read(id, buff, ofst)
}

func (f *FileSystem) Resize(path string, size int64) (err error) {
	fsFile := f.openLinkFile(path)
	err = fsFile.WriteSize(size)
	if err != nil {
		return err
	}
	id, err := fsFile.ReadId()
	if err != nil {
		return err
	}
	err = f.objectCacheSystem.Truncate(id, size)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileSystem) Rename(oldPath string, newPath string) (err error) {
	o := filepath.Join(f.fileSystemPath, oldPath)
	n := filepath.Join(f.fileSystemPath, newPath)
	err = os.Rename(o, n)
	if err != nil {
		return err
	}
	f.encryptQueue.Rename(oldPath, newPath)
	return nil
}

func GetRepoCtbRoot() (string, error) {
	root, err := os.Getwd()
	if err != nil {
		return "", err
	}
	path := filepath.Join(root, ".ctb")
	return path, nil
}

func (f *FileSystem) ObjectResolver(id string, writer io.Writer) (err error) {
	path := filepath.Join(f.ObjectPath, id)
	if _, err := os.Stat(path); os.IsNotExist(err) { //If object not exist, download it
		file, _ := os.Create(path)
		err = f.downloader.Download(id, file)
		if err != nil {
			return err
		}
	}
	file, _ := os.Open(path)
	decryptedReader, _ := f.fileCrypto.Decrypt(file, id)
	_, err = io.Copy(writer, decryptedReader)
	return
}

func (f *FileSystem) openLinkFile(path string) *link.Link {
	return link.New(path, f.fileSystemPath)
}
