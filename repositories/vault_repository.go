package repositories

import (
	"ctb-cli/core"
	"fmt"
	"os"
	"path/filepath"
)

// VaultRepository KeyStorePersist is an interface for persisting keys
type VaultRepository interface {
	GetVault(vaultId string) (core.Vault, error)
	SaveVault(vault core.Vault) (err error)
	AddKeyToVault(vault *core.Vault, keyIs string, serialized string) error
	GetKey(keyId string, vaultId string) (string, bool)
}

type VaultRepositoryFile struct {
	rootPath string
}

var _ VaultRepository = &VaultRepositoryFile{}

func NewVaultRepositoryFile(rootPath string) *VaultRepositoryFile {
	return &VaultRepositoryFile{
		rootPath: rootPath,
	}
}

func (k *VaultRepositoryFile) GetVault(vaultId string) (core.Vault, error) {
	p := filepath.Join(k.rootPath, vaultId)
	content, err := os.ReadFile(p)
	if err != nil {
		return core.Vault{}, err
	}
	return core.UnmarshalVault(content)
}

func (k *VaultRepositoryFile) SaveVault(vault core.Vault) (err error) {
	p := filepath.Join(k.rootPath, vault.Id)
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
	insidePath := k.vaultKeyFolder(vault.KeyId)
	err = os.MkdirAll(insidePath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (k *VaultRepositoryFile) GetKey(keyId string, vaultId string) (string, bool) {
	path := filepath.Join(k.vaultKeyFolder(vaultId), keyId)
	b, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(b), true
}

func (k *VaultRepositoryFile) AddKeyToVault(vault *core.Vault, keyId string, serialized string) error {
	path := filepath.Join(k.vaultKeyFolder(vault.KeyId), keyId)
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		return err
	}
	_, err = file.WriteString(serialized)
	if err != nil {
		return err
	}
	return nil
}

func (k *VaultRepositoryFile) vaultKeyFolder(vaultId string) string {
	return filepath.Join(k.rootPath, "K_"+vaultId)
}
