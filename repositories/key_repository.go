package repositories

import (
	"ctb-cli/core"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// KeyRepository KeyStorePersist is an interface for persisting keys
type KeyRepository interface {
	SaveDataKey(keyId, key, recipient string) error
	GetDataKey(keyID string, userId string) (string, error)
	GetPrivateKey(userId string) (string, error)
	SavePrivateKey(key string, userId string) (err error)
	GetVault(vaultId string) (core.Vault, error)
	SaveVault(vault core.Vault) (err error)
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

func (k *KeyRepositoryFile) GetPrivateKey(userId string) (string, error) {
	p := filepath.Join(k.getPrivatePath(), userId)
	content, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (k *KeyRepositoryFile) SavePrivateKey(key string, userId string) (err error) {
	p := filepath.Join(k.getPrivatePath(), userId)
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

func (k *KeyRepositoryFile) GetDataKey(keyID string, userId string) (string, error) {
	p := filepath.Join(k.getDataPath(userId), keyID)
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

func (k *KeyRepositoryFile) GetVault(vaultId string) (core.Vault, error) {
	p := filepath.Join(k.getVaultPath(), vaultId)
	content, err := os.ReadFile(p)
	if err != nil {
		return core.Vault{}, err
	}
	return core.UnmarshalVault(content)
}

func (k *KeyRepositoryFile) SaveVault(vault core.Vault) (err error) {
	p := filepath.Join(k.getVaultPath(), vault.Id)
	file, err := os.Create(p)
	if err != nil {
		return err
	}
	defer file.Close()
	serialized, err := vault.Marshal()
	if err != nil {
		return fmt.Errorf("error serializing vault")
	}
	_, err = file.Write(serialized)
	if err != nil {
		return err
	}
	return nil
}

func (k *KeyRepositoryFile) getVaultPath() string {
	p := filepath.Join(k.rootPath, "vault")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		os.MkdirAll(p, os.ModePerm)
	}
	return p
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
