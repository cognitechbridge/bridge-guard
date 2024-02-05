package encryptor

import (
	"io"
	"os"
)

type FileDecryptor struct {
	keystoreRepo KeystoreRepo
}

func NewFileDecryptor(keystoreRepo KeystoreRepo) FileDecryptor {
	return FileDecryptor{
		keystoreRepo: keystoreRepo,
	}
}

func (f *FileDecryptor) DecryptFile(file *os.File, fileId string) (read io.ReadCloser, err error) {
	key, err := f.keystoreRepo.Get(fileId)
	if err != nil {
		return nil, err
	}
	read, err = NewDecryptReader(key, file, file.Close)
	return read, err
}
