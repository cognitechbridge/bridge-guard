package object

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func (f *Service) StartEncryptRoutine() {
	for {
		item := <-f.encryptChan
		err := f.encrypt(item.id)
		if err != nil {
			continue
		}
	}
}

func (f *Service) encrypt(fileId string) (err error) {
	//Open object file
	inputFile, err := f.cache.AsFile(fileId)
	defer closeFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}

	//Create output file
	file, err := f.objectRepo.CreateFile(fileId)
	if err != nil {
		return fmt.Errorf("failed to Create output file: %w", err)
	}
	defer file.Close()

	//Create encrypted reader
	encryptedWriter, err := f.Encrypt(file, fileId)

	//Copy to output
	_, err = io.Copy(encryptedWriter, inputFile)
	if err != nil {
		return
	}
	err = encryptedWriter.Close()
	if err != nil {
		return
	}

	err = f.cache.Flush(fileId)
	if err != nil {
		return
	}
	fmt.Printf("File Encrypted: %s \n", fileId)

	path, _ := f.objectRepo.GetPath(fileId)

	f.uploadChan <- uploadChanItem{path: path}

	return nil
}

func (f *Service) StartUploadRoutine() {
	for {
		item := <-f.uploadChan
		err := f.upload(item.path)
		if err != nil {
			continue
		}
	}
}

func (f *Service) upload(path string) (err error) {
	_, fileId := filepath.Split(path)

	file, err := os.Open(path)
	if err != nil {
		return
	}

	//upload
	err = f.downloader.Upload(file, fileId)
	if err != nil {
		return
	}

	fmt.Printf("File Uploaded: %s \n", path)
	return nil
}
