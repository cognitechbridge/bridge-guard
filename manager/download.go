package manager

import (
	"ctb-cli/encryptor"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func (mn *Manager) Download(id string) error {

	//Create download file
	downloadPath := filepath.Join(mn.Filesystem.ObjectPath, id)
	downloadFile, err := mn.createFile(downloadPath)
	defer closeFile(downloadFile)
	if err != nil {
		return fmt.Errorf("error creating download file: %w", err)
	}

	//Download file
	err = mn.cloudStorage.Download(id, downloadFile)

	//Get data key
	dataKey, err := mn.store.Get(id)
	if err != nil {
		return fmt.Errorf("error creating FileDecryptor: %w", err)
	}

	// Create a new FileDecryptor
	rd, err := encryptor.NewFileDecryptor(*dataKey, downloadFile)
	if err != nil {
		return fmt.Errorf("error creating FileDecryptor: %w", err)
	}

	// Create or open the output file
	outputPath := filepath.Join(mn.Filesystem.ObjectCachePath, id)
	outputFile, err := mn.createFile(outputPath)
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

func (mn *Manager) createFile(path string) (*os.File, error) {
	parent := filepath.Dir(path)
	err := os.MkdirAll(parent, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return os.Create(path)
}
