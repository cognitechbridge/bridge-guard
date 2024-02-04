package filesyetem

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func (f *FileSystem) StartEncryptRoutine(input <-chan string) {
	for {
		path := <-input
		err := f.encrypt(path)
		if err != nil {
			continue
		}
	}
}

func (f *FileSystem) encrypt(path string) (err error) {
	fileId, err := f.GetFileId(path)
	absPath := filepath.Join(f.ObjectCachePath, fileId)

	//Open object file
	inputFile, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer closeFile(inputFile)

	//Create encrypted reader
	encryptedReader, err := f.encryptor.Encrypt(inputFile, fileId)

	//Create output file
	outPath := filepath.Join(f.ObjectPath, fileId)
	outFile, _ := os.Create(outPath)

	//Copy to output
	_, err = io.Copy(outFile, encryptedReader)
	if err != nil {
		return
	}

	fmt.Printf("File Encrypted: %s \n", path)
	return nil
}
