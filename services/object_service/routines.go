package object_service

import (
	"fmt"
	"io"
	"os"
)

// StartEncryptRoutine starts a routine that continuously encrypts items from the encryptChan channel.
// It calls the encrypt method for each item and continues to the next item if an error occurs.
func (o *Service) StartEncryptRoutine() {
	for {
		item := <-o.encryptChan
		err := o.encrypt(item)
		if err != nil {
			continue
		}
	}
}

// encrypt encrypts the object identified by the given ID using the provided encryption key.
// It opens the object file, creates an output file, and copies the encrypted content from the input file to the output file.
// After encrypting the file, it flushes the object from the cache and triggers an upload of the encrypted file.
// The function returns an error if any operation fails.
func (o *Service) encrypt(e encryptChanItem) (err error) {
	//Open object file
	inputFile, err := o.objectCacheRepo.AsFile(e.id)
	defer closeFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}

	//Create output file
	file, err := o.objectRepo.CreateFile(e.id, e.dir)
	if err != nil {
		return fmt.Errorf("failed to Create output file: %w", err)
	}
	defer file.Close()

	//Create encrypted writer
	encryptedWriter, err := o.encryptWriter(file, e.id, e.key)

	//Copy to output
	_, err = io.Copy(encryptedWriter, inputFile)
	if err != nil {
		return
	}
	//Close encrypted writer
	err = encryptedWriter.Close()
	if err != nil {
		return
	}
	//Flush the object from the cache
	err = o.objectCacheRepo.Flush(e.id)
	if err != nil {
		return
	}
	fmt.Printf("File Encrypted: %s \n", e.id)

	//Trigger upload
	o.uploadChan <- uploadChanItem{id: e.id, dir: e.dir}

	return nil
}

// StartUploadRoutine starts a routine that listens to the upload channel and processes the items.
// It continuously receives items from the upload channel and calls the upload method to handle each item.
// If an error occurs during the upload process, it will continue to the next item.
func (o *Service) StartUploadRoutine() {
	for {
		item := <-o.uploadChan
		err := o.upload(item.id, item.dir)
		if err != nil {
			continue
		}
	}
}

// upload uploads the file with the specified ID.
func (o *Service) upload(id string, objectPath string) error {
	// Get the dir of the object using the object repository
	path := o.objectRepo.GetPath(id, objectPath)
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening file for upload")
	}
	defer file.Close()
	// Upload the file
	err = o.downloader.Upload(file, id)
	if err != nil {
		return err
	}

	fmt.Printf("File Uploaded: %s \n", path)
	return nil
}
