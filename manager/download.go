package manager

import (
	"ctb-cli/encryptor"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Downloader struct {
	manger *Manager
	path   string
}

func (mn *Manager) NewDownloader(friendlyName string) *Downloader {
	return &Downloader{
		manger: mn,
		path:   friendlyName,
	}
}

func (dn *Downloader) Download() error {

	//Find file id
	id, err := dn.manger.Filesystem.GetFileId(dn.path)
	if err != nil {
		return fmt.Errorf("error creating FileDecryptor: %w", err)
	}

	//Create download file
	downloadPath := filepath.Join(dn.manger.Filesystem.ObjectPath, id)
	downloadFile, err := dn.createFile(downloadPath)
	defer closeFile(downloadFile)
	if err != nil {
		return fmt.Errorf("error creating download file: %w", err)
	}

	//Download file
	err = dn.manger.cloudStorage.Download(id, downloadFile)

	//Get data key
	dataKey, err := dn.manger.store.Get(id)
	if err != nil {
		return fmt.Errorf("error creating FileDecryptor: %w", err)
	}

	// Create a new FileDecryptor
	rd, err := encryptor.NewFileDecryptor(*dataKey, downloadFile)
	if err != nil {
		return fmt.Errorf("error creating FileDecryptor: %w", err)
	}

	// Create or open the output file
	outputPath := filepath.Join(dn.manger.Filesystem.ObjectCachePath, id)
	outputFile, err := dn.createFile(outputPath)
	defer closeFile(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}

	//Copy to output
	_, err = io.Copy(outputFile, rd)
	if err != nil {
		return fmt.Errorf("error Copying file: %w", err)
	}

	return nil
}

func (dn *Downloader) createFile(path string) (*os.File, error) {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return os.Create(path)
}
