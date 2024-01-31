package manager

import (
	"ctb-cli/encryptor"
	"fmt"
	"os"
	"path/filepath"
)

type Uploader struct {
	manger *Manager
}

func (mn *Manager) NewUploader() *Uploader {
	return &Uploader{
		manger: mn,
	}
}

func (dn *Uploader) UploadRoutine(input <-chan string) {
	for {
		path := <-input
		err := dn.upload(path)
		if err != nil {
			continue
		}
	}
}

func (dn *Uploader) upload(path string) (err error) {
	fileId, err := dn.manger.Filesystem.GetFileId(path)
	absPath := filepath.Join(dn.manger.Filesystem.ObjectCachePath, fileId)

	//Open object file
	inputFile, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer closeFile(inputFile)

	//Create header parameters
	clientId := dn.manger.config.ClientId
	pair, err := dn.manger.store.GenerateKeyPair(fileId)
	if err != nil {
		return
	}

	//Create reader
	efg := encryptor.NewFileEncryptor(
		inputFile,
		pair.Key,
		dn.manger.config.EncryptChunkSize,
		clientId,
		fileId,
		pair.RecoveryBlob,
	)

	//upload
	err = dn.manger.cloudStorage.Upload(efg, fileId)
	if err != nil {
		return
	}
	fmt.Printf("File Uploaded: %s \n", path)
	return nil
}
