package object_service

import (
	"ctb-cli/core"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

// encrypt encrypts the object identified by the given ID using the provided encryption key.
// It opens the object file, creates an output file, and copies the encrypted content from the input file to the output file.
// After encrypting the file, it flushes the object from the cache and triggers an upload of the encrypted file.
// The function returns an error if any operation fails.
func (o *Service) encrypt(link core.Link, key *core.KeyInfo) (err error) {
	//Open object file
	inputFile, err := o.objectCacheRepo.AsFile(link.Id())
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	//Create output file
	file, err := o.objectRepo.CreateFile(link)
	if err != nil {
		return fmt.Errorf("failed to Create output file: %w", err)
	}
	defer file.Close()

	//Create encrypted writer
	encryptedWriter, err := o.encryptWriter(file, link.Id(), key)
	if err != nil {
		return fmt.Errorf("failed to create encrypted writer: %w", err)
	}

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

	// Close object file
	err = file.Close()
	if err != nil {
		return
	}
	// Close cache file
	err = inputFile.Close()
	if err != nil {
		return
	}

	// Validate cache
	err = o.ValidateObject(link, key)
	if err != nil {
		return fmt.Errorf("Object validation failed: %w", err)
	}

	//Flush the object from the write cache
	err = o.objectCacheRepo.FlushFromWrite(link.Id())
	if err != nil {
		return
	}
	// Flush the object from the read cache
	err = o.objectCacheRepo.FlushFromRead(link.Id())
	if err != nil {
		return
	}

	log.Debugf("File Encrypted: %s", link.Id())
	fmt.Printf("File Encrypted: %s \n", link.Id())

	//Trigger upload
	//o.uploadChan <- uploadChanItem{id: link.Id(), path: link.Path}

	return nil
}

// StartUploadRoutine starts a routine that listens to the upload channel and processes the items.
// It continuously receives items from the upload channel and calls the upload method to handle each item.
// If an error occurs during the upload process, it will continue to the next item.
func (o *Service) StartUploadRoutine() {
	for {
		item := <-o.uploadChan
		err := o.upload(item.id, item.path)
		if err != nil {
			continue
		}
	}
}

// upload uploads the file with the specified ID.
func (o *Service) upload(id string, path string) error {
	// Get the dir of the object using the object repository
	objectPath := o.objectRepo.GetPath(id, path)
	// Open the file
	file, err := os.Open(objectPath)
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
