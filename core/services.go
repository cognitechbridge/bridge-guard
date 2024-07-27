package core

import (
	"io/fs"
)

type ObjectService interface {
	Read(link Link, buff []byte, ofst int64, key *KeyInfo) (int, error)
	Write(id string, buff []byte, ofst int64) (int, error)
	Create(id string) error
	Move(oldId string, newId string) error
	Truncate(id string, size int64) error
	GetKeyIdByObjectId(id string, dir string) (string, error)
	RemoveFromCache(id string) error
}

type FileSystemService interface {
	GetSubFiles(path string) (res []fs.FileInfo, err error)
	CreateFile(path string) (err error)
	CreateDir(path string) (err error)
	RemoveDir(path string) (err error)
	Write(path string, buff []byte, ofst int64) (n int, err error)
	Read(path string, buff []byte, ofst int64) (n int, err error)
	Rename(oldPath string, newPath string) (err error)
	RemovePath(path string) (err error)
	Resize(path string, size int64) (err error)
	Commit(path string) error
	OpenInWrite(path string) error
	GetUserFileAccess(path string, isDir bool) fs.FileMode
	GetDiskUsage() (totalBytes, freeBytes uint64, err error)
}

type KeyService interface {
	SetPrivateKey(privateKey PrivateKey)
	Get(keyID string, startVaultId string, startVaultPath string) (*KeyInfo, error)
	Insert(key *KeyInfo, path string) error
	Share(keyId string, startVaultId string, startVaultPath string, recipient PublicKey, recipientUserId string) error
	GetPublicKey() (PublicKey, error)
	GetPublicKeyByPrivateKey(PrivateKey PrivateKey) (PublicKey, error)
	CreateVault(parentId string, path string) (*Vault, error)
	GenerateKeyInVault(vaultId string, vaultPath string) (*KeyInfo, error)
	AddKeyToVault(vault *Vault, vaultPath string, key KeyInfo) error
	MoveVault(vaultId string, oldVaultPath string, newVaultPath string, oldParentVaultId string, oldParentVaultPath, newParentVaultId string, newParentVaultPath string) error
	MoveKey(keyId string, oldVaultId string, oldVaultPath string, newVaultId string, newVaultPath string) error
	GenerateUserKey() (*PrivateKey, error)
	IsUserJoined() bool
	GetHasAccessToKey(keyId string, startVaultId string, startVaultPath string, userId string) (bool, bool)
	GetKeyAccessList(keyId string, startVaultId string, startVaultPath string) (KeyAccessList, error)
	Unshare(keyId string, recipientUserId string, path string) error
}
