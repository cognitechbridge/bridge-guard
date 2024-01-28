package manager

import (
	"ctb-cli/encryptor"
	"fmt"
	"github.com/google/uuid"
	"os"
	"path/filepath"
)

type Uploader struct {
	manger       *Manager
	path         string
	friendlyName string
}

func (mn *Manager) NewUploader(path string, friendlyName string) *Uploader {
	return &Uploader{
		manger:       mn,
		path:         path,
		friendlyName: friendlyName,
	}
}

func (dn *Uploader) Upload() (string, error) {
	absPath, _ := filepath.Abs(dn.path)

	//Open input file
	inputFile, err := os.Open(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to open input file: %w", err)
	}
	defer closeFile(inputFile)

	//Create output file
	outputFile, err := os.Create("D:\\encrypted.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer closeFile(outputFile)

	//Create header parameters
	clientId := dn.manger.config.ClientId
	fileUuid, _ := uuid.NewV7()
	pair, err := dn.manger.store.GenerateKeyPair(fileUuid.String())
	if err != nil {
		return "", err
	}

	//Create reader
	efg := encryptor.NewFileEncryptor(
		inputFile,
		pair.Key,
		dn.manger.config.EncryptChunkSize,
		clientId,
		fileUuid.String(),
		pair.RecoveryBlob,
	)

	//Save friendly name
	err = dn.manger.filesystem.SavePath(fileUuid.String(), dn.friendlyName)
	if err != nil {
		return "", err
	}

	//Upload
	//err = dn.manger.s3storage.Upload(efg, fileUuid.String())
	err = dn.manger.cloudStorage.Upload(efg, fileUuid.String())
	if err != nil {
		return "", err
	}

	//Return file id
	return fileUuid.String(), nil
}
