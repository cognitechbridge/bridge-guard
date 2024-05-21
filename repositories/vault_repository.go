package repositories

import (
	"ctb-cli/core"
	"encoding/json"
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
	GetVaultParent(vaultPath string) (string, core.Vault, error)
	GetVaultByPath(path string) (core.Vault, error)
	RemoveVaultLink(path string) error
	GetFileVault(path string) (core.Vault, string, error)
}

type VaultRepositoryFile struct {
	rootPath string
}

type vaultLink struct {
	VaultId string `json:"vaultId"`
}

func NewVaultLink(vaultId string, keyId string) vaultLink {
	return vaultLink{
		VaultId: vaultId,
	}
}

var _ VaultRepository = &VaultRepositoryFile{}

func NewVaultRepositoryFile(rootPath string) *VaultRepositoryFile {
	return &VaultRepositoryFile{
		rootPath: rootPath,
	}
}

func (k *VaultRepositoryFile) GetVault(vaultId string, vaultPath string) (core.Vault, error) {
	p := k.vaultFile(vaultId, vaultPath)
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
	// Insert Vault Link
	link := vaultLink{
		VaultId: vault.Id,
	}
	err = k.insertVaultLink(vaultPath, link)
	if err != nil {
		return err
	}
	return nil
}

func (k *VaultRepositoryFile) SaveVault(vault core.Vault, vaultPath string) (err error) {
	p := k.vaultFile(vault.Id, vaultPath)
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

func (k *VaultRepositoryFile) GetVaultParent(vaultPath string) (string, core.Vault, error) {
	if filepath.Clean(vaultPath) == string(filepath.Separator) {
		return "", core.Vault{}, nil
	}
	parentPath := filepath.Dir(vaultPath)
	parentLink, err := k.GetVaultByPath(parentPath)
	if err != nil {
		return "", core.Vault{}, err
	}
	return parentPath, parentLink, nil
}

// GetVaultByPath retrieves the vault by the given path.
// It reads the vault link file located at the specified path and returns the corresponding Vault object.
// If an error occurs during file reading or unmarshaling, it returns an empty Vault object and the error.
func (k *VaultRepositoryFile) GetVaultByPath(path string) (core.Vault, error) {
	// Read vault link file
	link, err := k.getVaultLinkByPath(path)
	if err != nil {
		return core.Vault{}, err
	}
	// Read vault file
	vault, err := k.GetVault(link.VaultId, path)
	if err != nil {
		return core.Vault{}, err
	}
	return vault, nil
}

// GetVaultLinkByPath retrieves the vault link by the given path.
// It reads the vault link file located at the specified path and returns the corresponding VaultLink object.
// If an error occurs during file reading or unmarshaling, it returns an empty VaultLink object and the error.
func (k *VaultRepositoryFile) getVaultLinkByPath(path string) (vaultLink, error) {
	// Read the vault link file
	p := k.getVaultLinkPath(path)
	js, err := os.ReadFile(p)
	if err != nil {
		return vaultLink{}, fmt.Errorf("error reading vault link file: %v", err)
	}
	var link vaultLink
	err = json.Unmarshal(js, &link)
	if err != nil {
		return vaultLink{}, fmt.Errorf("error unmarshalink vault link file: %v", err)
	}
	return link, nil
}

// RemoveVaultLink removes the vault link file for the specified path.
// It takes the path of the link file as input and returns an error if any.
func (k *VaultRepositoryFile) RemoveVaultLink(path string) error {
	absPath := k.getVaultLinkPath(path)
	err := os.Remove(absPath)
	if err != nil {
		return ErrRemovingVaultLinkFile
	}
	return nil
}

// getVaultLink retrieves the vault link for the file located at the specified path.
// It takes a path string as input and returns a core.VaultLink and an error.
func (k *VaultRepositoryFile) GetFileVault(path string) (core.Vault, string, error) {
	dir := filepath.Dir(path)
	vaultLink, err := k.GetVaultByPath(dir)
	return vaultLink, dir, err
}

// InsertVaultLink inserts a VaultLink into the specified path.
// It creates the necessary directories and writes the link data to a file.
// The link data is serialized as JSON before writing to the file.
// If any error occurs during the process, it is returned.
func (k *VaultRepositoryFile) insertVaultLink(path string, link vaultLink) error {
	absPath := k.getVaultLinkPath(path)
	err := os.MkdirAll(filepath.Dir(absPath), os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(absPath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	js, _ := json.Marshal(link)
	_, _ = file.Write(js)
	return nil
}

// vaultFolder returns the path to the vault folder for the specified path.
func (k *VaultRepositoryFile) vaultFolder(vaultPath string) string {
	return filepath.Join(k.rootPath, vaultPath, ".meta", ".vault")
}

// getVaultLinkPath returns the path to the vault link file for the specified path.
func (k *VaultRepositoryFile) getVaultLinkPath(path string) string {
	return filepath.Join(k.vaultFolder(path), ".link")
}

// vaultKeyFolder returns the path to the key folder for the specified vault ID and path.
func (k *VaultRepositoryFile) vaultKeyFolder(vaultId string, vaultPath string) string {
	return filepath.Join(k.vaultFolder(vaultPath), "."+vaultId)
}

// vaultFile returns the path to the vault file for the specified vault ID and path.
func (k *VaultRepositoryFile) vaultFile(vaultId string, vaultPath string) string {
	return filepath.Join(k.vaultFolder(vaultPath), vaultId)
}
