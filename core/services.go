package core

import "io/fs"

type ObjectService interface {
	Read(id string, buff []byte, ofst int64, key *Key) (int, error)
	Write(id string, buff []byte, ofst int64) (int, error)
	Create(id string) error
	Move(oldId string, newId string) error
	Truncate(id string, size int64) error
	GetKeyIdByObjectId(id string) (string, error)
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
}

type KeyService interface {
	Get(keyID string, startVaultId string) (*Key, error)
	Insert(key *KeyInfo) error
	GetRecoveryItems() ([]RecoveryItem, error)
	AddRecoveryKey(inPath string) error
	GenerateUserKeys() (err error)
	SetSecret(secret string)
	LoadKeys() error
	ChangeSecret(secret string) error
	Share(keyId string, recipient []byte, recipientUserId string) error
	GetPublicKey() ([]byte, error)
	SetUserId(userId string)
	CreateVault(parentId string) (*Vault, error)
	GenerateKeyInVault(vaultId string) (*KeyInfo, error)
	AddKeyToVault(vault *Vault, key KeyInfo) error
	MoveVault(vaultLink VaultLink, oldVault VaultLink, newVault VaultLink) error
	MoveKey(keyId string, oldVault VaultLink, newVault VaultLink) error
}
