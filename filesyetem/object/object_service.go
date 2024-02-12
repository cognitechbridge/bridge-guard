package object

import (
	"ctb-cli/crypto/file_crypto"
	"ctb-cli/crypto/recovery"
	"ctb-cli/types"
	"io"
)

type Service struct {
	keystoreRepo keystoreRepo
	clientId     string
}

type keystoreRepo interface {
	Get(keyID string) (*types.Key, error)
	Insert(keyID string, key types.Key) error
	GetRecoveryItems() ([]types.RecoveryItem, error)
}

func NewService(keystoreRepo keystoreRepo, clientId string) Service {
	return Service{
		keystoreRepo: keystoreRepo,
		clientId:     clientId,
	}
}

func (f *Service) Encrypt(writer io.Writer, fileId string) (write io.WriteCloser, err error) {
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
	return file_crypto.NewWriter(writer, pair.Key, f.clientId, fileId, pair.RecoveryBlobs)
}

func (f *Service) Decrypt(reader io.Reader, fileId string) (read io.Reader, err error) {
	key, err := f.keystoreRepo.Get(fileId)
	if err != nil {
		return nil, err
	}
	read, err = file_crypto.NewReader(key, reader)
	return read, err
}
