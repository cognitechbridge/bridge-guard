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
	p := filepath.Join(k.getVaultPath(), vaultId)
	content, err := os.ReadFile(p)
	if err != nil {
		return core.Vault{}, err
	}
	return core.UnmarshalVault(content)
}

func (k *VaultRepositoryFile) SaveVault(vault core.Vault) (err error) {
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

func (k *VaultRepositoryFile) getVaultPath() string {
	p := filepath.Join(k.rootPath, "vault")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		os.MkdirAll(p, os.ModePerm)
	}
	return p
}
