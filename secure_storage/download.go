package secure_storage

import (
	"ctb-cli/encryptor"
	"ctb-cli/utils"
	"fmt"
	"io"
	"os"
)

type Downloader struct {
	manger       *Manager
	path         string
	friendlyName string
}

func (mn *Manager) NewDownloader(path string, friendlyName string) *Downloader {
	return &Downloader{
		manger:       mn,
		path:         path,
		friendlyName: friendlyName,
	}
}

func (dn *Downloader) Download() error {

	//Find file id
	id, err := dn.manger.filesystem.GetPath(dn.friendlyName)
	if err != nil {
		return fmt.Errorf("error creating FileDecryptor: %w", err)
	}

	//Create temp download file
	tempFile, err := utils.CreateTempFile(id)
	if err != nil {
		return fmt.Errorf("error creating FileDecryptor: %w", err)
	}
	defer utils.CloseDeleteTempFile(tempFile)

	//Download file
	err = dn.manger.cloudStorage.Download(id, tempFile)

	//Get data key
	dataKey, err := dn.manger.store.Get(id)
	if err != nil {
		return fmt.Errorf("error creating FileDecryptor: %w", err)
	}

	// Create a new FileDecryptor
	rd, err := encryptor.NewFileDecryptor(*dataKey, tempFile)
	if err != nil {
		return fmt.Errorf("error creating FileDecryptor: %w", err)
	}

	// Create or open the output file
	outputFile, err := os.Create(dn.path) // Specify your output file path
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer closeFile(outputFile)

	//Copy to output
	_, err = io.Copy(outputFile, rd)
	if err != nil {
		return fmt.Errorf("error Copying file: %w", err)
	}

	return nil
}
