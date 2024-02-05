package encryptor

import (
	"io"
)

type FileDecryptor struct {
	keystoreRepo KeystoreRepo
}

func NewFileDecryptor(keystoreRepo KeystoreRepo) FileDecryptor {
	return FileDecryptor{
		keystoreRepo: keystoreRepo,
	}
}

func (f *FileDecryptor) Decrypt(reader io.Reader, fileId string) (read io.Reader, err error) {
	key, err := f.keystoreRepo.Get(fileId)
	if err != nil {
		return nil, err
	}
	read, err = NewDecryptReader(key, reader)
	return read, err
}
