package secure_storage

import (
	"ctb-cli/encryptor"
	"fmt"
	"github.com/google/uuid"
	"os"
)

type Uploader struct {
	manger       *Manager
	path         string
	friendlyName string
	clientId     string
}

func (mn *Manager) NewUploader(path string, friendlyName string, clientId string) *Uploader {
	return &Uploader{
		manger:       mn,
		path:         path,
		friendlyName: friendlyName,
		clientId:     clientId,
	}
}

func (dn *Uploader) Upload() (string, error) {
	//Open input file
	inputFile, err := os.Open(dn.path)
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
	clientId := dn.clientId
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
