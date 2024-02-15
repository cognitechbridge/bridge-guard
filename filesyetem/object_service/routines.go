package object_service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func (o *Service) StartEncryptRoutine() {
	for {
		item := <-o.encryptChan
		err := o.encrypt(item.id)
		if err != nil {
			continue
		}
	}
}

func (o *Service) encrypt(fileId string) (err error) {
	//Open object file
	inputFile, err := o.cache.AsFile(fileId)
	defer closeFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}

	//Create output file
	file, err := o.objectRepo.CreateFile(fileId)
	if err != nil {
		return fmt.Errorf("failed to Create output file: %w", err)
	}
	defer file.Close()

	//Create encrypted writer
	encryptedWriter, err := o.encryptWriter(file, fileId)

	//Copy to output
	_, err = io.Copy(encryptedWriter, inputFile)
	if err != nil {
		return
	}
	err = encryptedWriter.Close()
	if err != nil {
		return
	}

	err = o.cache.Flush(fileId)
	if err != nil {
		return
	}
	fmt.Printf("File Encrypted: %s \n", fileId)

	path, _ := o.objectRepo.GetPath(fileId)

	o.uploadChan <- uploadChanItem{path: path}

	return nil
}

func (o *Service) StartUploadRoutine() {
	for {
		item := <-o.uploadChan
		err := o.upload(item.path)
		if err != nil {
			continue
		}
	}
}

func (o *Service) upload(path string) (err error) {
	_, fileId := filepath.Split(path)

	file, err := os.Open(path)
	if err != nil {
		return
	}

	//upload
	err = o.downloader.Upload(file, fileId)
	if err != nil {
		return
	}

	fmt.Printf("File Uploaded: %s \n", path)
	return nil
}
