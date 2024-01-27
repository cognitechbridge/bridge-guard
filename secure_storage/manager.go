package secure_storage

import (
	"storage-go/filesyetem"
	"storage-go/keystore"
	"storage-go/persist_file"
)

type Manager struct {
	store        *keystore.KeyStore
	cloudStorage persist_file.CloudStorageClient
	filesystem   *filesyetem.FileSystem
	config       ManagerConfig
}

type ManagerConfig struct {
	EncryptChunkSize uint64
}

func NewManager(
	config ManagerConfig,
	keyStore *keystore.KeyStore,
	filesyetem *filesyetem.FileSystem,
	cloudStorage persist_file.CloudStorageClient,
) *Manager {
	return &Manager{
		store:        keyStore,
		filesystem:   filesyetem,
		cloudStorage: cloudStorage,
		config:       config,
	}
}
