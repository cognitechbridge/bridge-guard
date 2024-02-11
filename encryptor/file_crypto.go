package encryptor

import (
	"ctb-cli/encryptor/recovery"
	"ctb-cli/types"
	"io"
)

type FileEncryptor struct {
	keystoreRepo KeystoreRepo
	clientId     string
}

type KeystoreRepo interface {
	Get(keyID string) (*Key, error)
	Insert(keyID string, key Key) error
	GetRecoveryItems() ([]types.RecoveryItem, error)
}

func NewFileEncryptor(keystoreRepo KeystoreRepo, clientId string) FileEncryptor {
	return FileEncryptor{
		keystoreRepo: keystoreRepo,
		clientId:     clientId,
	}
}

func (f *FileEncryptor) Encrypt(writer io.Writer, fileId string) (write io.WriteCloser, err error) {
	recoveryItems, err := f.keystoreRepo.GetRecoveryItems()
	if err != nil {
		return nil, err
	}
	pair, err := recovery.GenerateKeyPair(recoveryItems)
	if err != nil {
		return nil, err
	}
	err = f.keystoreRepo.Insert(fileId, pair.Key)
	if err != nil {
		return nil, err
	}
	return NewWriter(writer, pair.Key, f.clientId, fileId, pair.RecoveryBlobs)
}

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
	read, err = NewReader(key, reader)
	return read, err
}
