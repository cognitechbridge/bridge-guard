package encryptor

import (
	"ctb-cli/keystore"
	"io"
)

type FileEncryptor struct {
	keystoreRepo KeystoreRepo
	chunkSize    uint64
	clientId     string
}

type KeystoreRepo interface {
	GenerateKeyPair(keyId string) (keystore.GeneratedKey, error)
	Get(keyID string) (*Key, error)
}

func NewFileEncryptor(keystoreRepo KeystoreRepo, chunkSize uint64, clientId string) FileEncryptor {
	return FileEncryptor{
		keystoreRepo: keystoreRepo,
		chunkSize:    chunkSize,
		clientId:     clientId,
	}
}

func (f *FileEncryptor) Encrypt(reader io.Reader, fileId string) (read io.Reader, err error) {
	pair, err := f.keystoreRepo.GenerateKeyPair(fileId)
	if err != nil {
		return nil, err
	}
	return NewEncryptReader(reader, pair.Key, f.chunkSize, f.clientId, fileId, pair.RecoveryBlobs), nil
}
