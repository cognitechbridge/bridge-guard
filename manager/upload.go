package manager

import (
	"ctb-cli/encryptor"
	"fmt"
	"os"
	"path/filepath"
)

func (mn *Manager) UploadRoutine(input <-chan string) {
	for {
		path := <-input
		err := mn.upload(path)
		if err != nil {
			continue
		}
	}
}

func (mn *Manager) upload(path string) (err error) {
	fileId, err := mn.Filesystem.GetFileId(path)
	absPath := filepath.Join(mn.Filesystem.ObjectCachePath, fileId)

	//Open object file
	inputFile, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer closeFile(inputFile)

	//Create header parameters
	clientId := mn.config.ClientId
	pair, err := mn.store.GenerateKeyPair(fileId)
	if err != nil {
		return
	}

	//Create reader
	efg := encryptor.NewEncryptReader(
		inputFile,
		pair.Key,
		mn.config.EncryptChunkSize,
		clientId,
		fileId,
		pair.RecoveryBlob,
	)

	//upload
	err = mn.cloudStorage.Upload(efg, fileId)
	if err != nil {
		return
	}
	fmt.Printf("File Uploaded: %s \n", path)
	return nil
}
