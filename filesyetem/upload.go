package filesyetem

import (
	"fmt"
	"os"
	"path/filepath"
)

func (f *FileSystem) UploadRoutine() {
	for {
		item := <-f.uploadChan
		err := f.upload(item.path)
		if err != nil {
			continue
		}
	}
}

func (f *FileSystem) upload(path string) (err error) {
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
