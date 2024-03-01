package object_service

import (
	"fmt"
	"io"
	"os"
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
	inputFile, err := o.objectCacheRepo.AsFile(fileId)
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

	err = o.objectCacheRepo.Flush(fileId)
	if err != nil {
		return
	}
	fmt.Printf("File Encrypted: %s \n", fileId)

	o.uploadChan <- uploadChanItem{id: fileId}

	return nil
}

func (o *Service) StartUploadRoutine() {
	for {
		item := <-o.uploadChan
		err := o.upload(item.id)
		if err != nil {
			continue
		}
	}
}

func (o *Service) upload(id string) error {
	path, err := o.objectRepo.GetPath(id)
	if err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening file for upload")
	}
	//upload
	err = o.downloader.Upload(file, id)
	if err != nil {
		return err
	}

	fmt.Printf("File Uploaded: %s \n", path)
	return nil
}
