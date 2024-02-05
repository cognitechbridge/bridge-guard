package filesyetem

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func (f *FileSystem) StartEncryptRoutine() {
	for {
		item := <-f.encryptChan
		err := f.encrypt(item.id)
		if err != nil {
			continue
		}
	}
}

func (f *FileSystem) encrypt(fileId string) (err error) {
	//Open object file
	inputFile, err := f.objectCacheSystem.AsFile(fileId)
	defer closeFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}

	//Create encrypted reader
	encryptedReader, err := f.encryptor.Encrypt(inputFile, fileId)

	//Create output file
	outPath := filepath.Join(f.ObjectPath, fileId)
	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to Create output file: %w", err)
	}
	defer outFile.Close()

	//Copy to output
	_, err = io.Copy(outFile, encryptedReader)
	if err != nil {
		return
	}

	err = f.objectCacheSystem.Flush(fileId)
	if err != nil {
		return
	}
	fmt.Printf("File Encrypted: %s \n", fileId)

	f.UploadChan <- outPath

	return nil
}
