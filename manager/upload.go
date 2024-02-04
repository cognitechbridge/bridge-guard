package manager

import (
	"fmt"
	"os"
	"path/filepath"
)

func (mn *Manager) UploadRoutine(input <-chan string) {
	for {
		path := <-input
		err := mn.upload(path)
		if err != nil {
			continue
		}
	}
}

func (mn *Manager) upload(path string) (err error) {
	_, fileId := filepath.Split(path)

	file, err := os.Open(path)
	if err != nil {
		return
	}

	//upload
	err = mn.cloudStorage.Upload(file, fileId)
	if err != nil {
		return
	}

	fmt.Printf("File Uploaded: %s \n", path)
	return nil
}
