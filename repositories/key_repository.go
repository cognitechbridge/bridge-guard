package repositories

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

var (
	KeyNotFound = errors.New("key not found")
)

// KeyRepository KeyStorePersist is an interface for persisting keys
type KeyRepository interface {
	SaveDataKey(keyId, key, recipient string) error
	GetDataKey(keyID string, userId string) (string, error)
	DataKeyExist(keyId string, userId string) bool
}

type KeyRepositoryFile struct {
	rootPath string
}

var _ KeyRepository = &KeyRepositoryFile{}

func NewKeyRepositoryFile(rootPath string) *KeyRepositoryFile {
	return &KeyRepositoryFile{
		rootPath: rootPath,
	}
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

func (k *KeyRepositoryFile) GetDataKey(keyID string, userId string) (string, error) {
	p := filepath.Join(k.getDataPath(userId), keyID)
	file, err := os.Open(p)
	defer file.Close()
	if os.IsNotExist(err) {
		return "", KeyNotFound
	}
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
	p := filepath.Join(k.rootPath, "data", recipient)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		os.MkdirAll(p, os.ModePerm)
	}
	return p
}

func (k *KeyRepositoryFile) getPrivatePath() string {
	p := filepath.Join(k.rootPath, "private")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		os.MkdirAll(p, os.ModePerm)
	}
	return p
}

func (k *KeyRepositoryFile) DataKeyExist(keyId string, userId string) bool {
	p := filepath.Join(k.getDataPath(userId), keyId)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}
