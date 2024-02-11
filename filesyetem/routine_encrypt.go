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

	//Create output file
	outPath := filepath.Join(f.ObjectPath, fileId)
	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to Create output file: %w", err)
	}
	defer outFile.Close()

	//Create encrypted reader
	encryptedWriter, err := f.fileCrypto.Encrypt(outFile, fileId)

	//Copy to output
	_, err = io.Copy(encryptedWriter, inputFile)
	if err != nil {
		return
	}
	err = encryptedWriter.Close()
	if err != nil {
		return
	}

	err = f.objectCacheSystem.Flush(fileId)
	if err != nil {
		return
	}
	fmt.Printf("File Encrypted: %s \n", fileId)

	f.uploadChan <- uploadChanItem{path: outPath}

	return nil
}
