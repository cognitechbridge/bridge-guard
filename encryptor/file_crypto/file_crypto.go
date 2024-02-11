package file_crypto

import (
	"ctb-cli/encryptor/recovery"
	"ctb-cli/types"
	"io"
)

type Key = types.Key

type FileCrypto struct {
	keystoreRepo KeystoreRepo
	clientId     string
}

type KeystoreRepo interface {
	Get(keyID string) (*Key, error)
	Insert(keyID string, key Key) error
	GetRecoveryItems() ([]types.RecoveryItem, error)
}

func New(keystoreRepo KeystoreRepo, clientId string) FileCrypto {
	return FileCrypto{
		keystoreRepo: keystoreRepo,
		clientId:     clientId,
	}
}

func (f *FileCrypto) Encrypt(writer io.Writer, fileId string) (write io.WriteCloser, err error) {
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
	return newWriter(writer, pair.Key, f.clientId, fileId, pair.RecoveryBlobs)
}

func (f *FileCrypto) Decrypt(reader io.Reader, fileId string) (read io.Reader, err error) {
	key, err := f.keystoreRepo.Get(fileId)
	if err != nil {
		return nil, err
	}
	read, err = newReader(key, reader)
	return read, err
}
