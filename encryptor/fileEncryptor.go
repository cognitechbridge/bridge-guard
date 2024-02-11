package encryptor

import (
	"ctb-cli/encryptor/recovery"
	"ctb-cli/types"
	"io"
)

type FileEncryptor struct {
	keystoreRepo KeystoreRepo
	chunkSize    uint64
	clientId     string
}

type KeystoreRepo interface {
	Get(keyID string) (*Key, error)
	Insert(keyID string, key Key) error
	GetRecoveryItems() ([]types.RecoveryItem, error)
}

func NewFileEncryptor(keystoreRepo KeystoreRepo, chunkSize uint64, clientId string) FileEncryptor {
	return FileEncryptor{
		keystoreRepo: keystoreRepo,
		chunkSize:    chunkSize,
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
	return NewWriter(writer, pair.Key, f.chunkSize, f.clientId, fileId, pair.RecoveryBlobs)
}
