package repositories

import (
	"io"
	"os"
	"path/filepath"
)

// KeyRepository KeyStorePersist is an interface for persisting keys
type KeyRepository interface {
	SaveDataKey(keyId, key, recipient string) error
	GetDataKey(keyID string) (string, error)
	GetPrivateKey() (string, error)
	SavePrivateKey(key string) (err error)
	SavePublicKey(id string, key string) (err error)
}

type KeyRepositoryFile struct {
	clientId string
	rootPath string
}

var _ KeyRepository = &KeyRepositoryFile{}

func NewKeyRepositoryFile(clientId string, rootPath string) *KeyRepositoryFile {
	return &KeyRepositoryFile{
		rootPath: rootPath,
		clientId: clientId,
	}
}

func (k *KeyRepositoryFile) SavePublicKey(id string, key string) (err error) {
	p := filepath.Join(k.getPublicPath(), id)

	file, err := os.Create(p)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(key)
	if err != nil {
		return err
	}
	return nil
}

func (k *KeyRepositoryFile) GetPrivateKey() (string, error) {
	p := filepath.Join(k.getPrivatePath(), k.clientId)
	content, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (k *KeyRepositoryFile) SavePrivateKey(key string) (err error) {
	p := filepath.Join(k.getPrivatePath(), k.clientId)
	file, err := os.Create(p)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(key)
	if err != nil {
		return err
	}
	return nil
}

func (k *KeyRepositoryFile) SaveDataKey(keyId, key, recipient string) error {
	p := filepath.Join(k.getDataPath(recipient), keyId)
	file, err := os.Create(p)
	defer file.Close()
	if err != nil {
		return err
	}
	_, err = file.Write([]byte(key))
	if err != nil {
		return err
	}
	return nil
}

func (k *KeyRepositoryFile) GetDataKey(keyID string) (string, error) {
	p := filepath.Join(k.getDataPath(k.clientId), keyID)
	file, err := os.Open(p)
	defer file.Close()
	if err != nil {
		return "", err
	}
	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(content), err
}

func (k *KeyRepositoryFile) getDataPath(recipient string) string {
	p := filepath.Join(k.rootPath, "keys", "data", recipient)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		os.MkdirAll(p, os.ModePerm)
	}
	return p
}

func (k *KeyRepositoryFile) getPrivatePath() string {
	p := filepath.Join(k.rootPath, "keys", "private")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		os.MkdirAll(p, os.ModePerm)
	}
	return p
}

func (k *KeyRepositoryFile) getPublicPath() string {
	p := filepath.Join(k.rootPath, "keys", "public")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		os.MkdirAll(p, os.ModePerm)
	}
	return p
}
