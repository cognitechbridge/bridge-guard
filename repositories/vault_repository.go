package repositories

import (
	"ctb-cli/core"
	"fmt"
	"os"
	"path/filepath"
)

// VaultRepository KeyStorePersist is an interface for persisting keys
type VaultRepository interface {
	GetVault(vaultId string, vaultPath string) (core.Vault, error)
	SaveVault(vault core.Vault, vaultPath string) (err error)
	InsertVault(vault core.Vault, vaultPath string) error
	AddKeyToVault(vault *core.Vault, vaultPath string, keyId string, serialized string) error
	GetKey(keyId string, vaultId string, vaultPath string) (string, bool)
	RemoveKey(keyId string, vaultId string, vaultPath string) error
	GetVaultParentPath(vaultPath string) string
	MoveVault(oldVaultPath string, newVaultPath string) error
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

func (k *VaultRepositoryFile) GetVault(vaultId string, vaultPath string) (core.Vault, error) {
	p := filepath.Join(k.rootPath, vaultPath, vaultId)
	content, err := os.ReadFile(p)
	if err != nil {
		return core.Vault{}, err
	}
	return core.UnmarshalVault(content)
}

func (k *VaultRepositoryFile) InsertVault(vault core.Vault, vaultPath string) error {
	err := k.SaveVault(vault, vaultPath)
	if err != nil {
		return err
	}
	insidePath := k.vaultKeyFolder(vault.Id, vaultPath)
	err = os.MkdirAll(insidePath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (k *VaultRepositoryFile) SaveVault(vault core.Vault, vaultPath string) (err error) {
	p := filepath.Join(k.rootPath, vaultPath, vault.Id)
	err = os.MkdirAll(filepath.Dir(p), os.ModePerm)
	if err != nil {
		return err
	}
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
	insidePath := k.vaultKeyFolder(vault.Id, vaultPath)
	err = os.MkdirAll(insidePath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (k *VaultRepositoryFile) GetKey(keyId string, vaultId string, vaultPath string) (string, bool) {
	path := filepath.Join(k.vaultKeyFolder(vaultId, vaultPath), keyId)
	b, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(b), true
}

func (k *VaultRepositoryFile) AddKeyToVault(vault *core.Vault, vaultPath string, keyId string, serialized string) error {
	path := filepath.Join(k.vaultKeyFolder(vault.Id, vaultPath), keyId)
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(serialized)
	if err != nil {
		return err
	}
	return nil
}

func (k *VaultRepositoryFile) RemoveKey(keyId string, vaultId string, vaultPath string) error {
	path := filepath.Join(k.vaultKeyFolder(vaultId, vaultPath), keyId)
	return os.Remove(path)
}

func (k *VaultRepositoryFile) GetVaultParentPath(vaultPath string) string {
	return filepath.Dir(vaultPath)
}

func (k *VaultRepositoryFile) MoveVault(oldVaultPath string, newVaultPath string) error {
	oldPath := filepath.Join(k.rootPath, oldVaultPath)
	newPath := filepath.Join(k.rootPath, newVaultPath)
	return os.Rename(oldPath, newPath)
}

func (k *VaultRepositoryFile) vaultKeyFolder(vaultId string, vaultPath string) string {
	return filepath.Join(k.rootPath, vaultPath, "K_"+vaultId)
}
