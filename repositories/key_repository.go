package repositories

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

var (
	ErrKeyNotFound   = errors.New("key not found")
	ErrUserNotJoined = errors.New("user not joined")
)

// KeyRepository KeyStorePersist is an interface for persisting keys
type KeyRepository interface {
	SaveDataKey(keyId, key, recipient string) error
	GetDataKey(keyID string, userId string) (string, error)
	DataKeyExist(keyId string, userId string) bool
	IsUserJoined(userId string) bool
	JoinUser(userId string) error
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
	datapath, err := k.getDataPath(recipient)
	if err != nil {
		return err
	}
	p := filepath.Join(datapath, keyId)
	file, err := os.Create(p)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write([]byte(key))
	if err != nil {
		return err
	}
	return nil
}

func (k *KeyRepositoryFile) GetDataKey(keyID string, userId string) (string, error) {
	datapath, err := k.getDataPath(userId)
	if err != nil {
		return "", err
	}
	p := filepath.Join(datapath, keyID)
	file, err := os.Open(p)
	if os.IsNotExist(err) {
		return "", ErrKeyNotFound
	}
	if err != nil {
		return "", err
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(content), err
}

func (k *KeyRepositoryFile) getDataPath(recipient string) (string, error) {
	p := filepath.Join(k.rootPath, "data", recipient)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return "", ErrUserNotJoined
	}
	return p, nil
}

func (k *KeyRepositoryFile) DataKeyExist(keyId string, userId string) bool {
	datapath, err := k.getDataPath(userId)
	if err != nil {
		return false
	}
	p := filepath.Join(datapath, keyId)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func (k *KeyRepositoryFile) IsUserJoined(userId string) bool {
	p := filepath.Join(k.rootPath, "data", userId)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

// JoinUser creates a directory for the specified user ID in the key repository.
// It takes the user ID as a parameter and returns an error if any.
func (k *KeyRepositoryFile) JoinUser(userId string) error {
	p := filepath.Join(k.rootPath, "data", userId)
	os.MkdirAll(p, os.ModePerm)
	return nil
}
