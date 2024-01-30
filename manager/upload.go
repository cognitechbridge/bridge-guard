package manager

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
	isDir        bool
	force        bool
}

func (mn *Manager) NewUploader(path string, friendlyName string, isDir bool, force bool) *Uploader {
	return &Uploader{
		manger:       mn,
		path:         path,
		isDir:        isDir,
		friendlyName: friendlyName,
		force:        force,
	}
}

func (dn *Uploader) Upload() (string, error) {
	//Check if path already exist
	exist, _ := dn.manger.Filesystem.PathExist(dn.friendlyName)
	if exist {
		if !dn.force {
			return "", fmt.Errorf("file exist: %s", dn.friendlyName)
		} else {
			if err := dn.manger.Filesystem.RemovePath(dn.friendlyName); err != nil {
				return "", err
			}
		}
	}

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
	err = dn.manger.Filesystem.CreateFsFile(fileUuid.String(), dn.friendlyName, 0)
	if err != nil {
		return "", err
	}

	//Upload
	err = dn.manger.cloudStorage.Upload(efg, fileUuid.String(), dn.friendlyName)
	if err != nil {
		return "", err
	}

	//Return file id
	return fileUuid.String(), nil
}
