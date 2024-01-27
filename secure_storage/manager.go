package secure_storage

import (
	"ctb-cli/file_db"
	"ctb-cli/filesyetem"
	"ctb-cli/keystore"
)

type Manager struct {
	store        *keystore.KeyStore
	cloudStorage file_db.CloudStorageClient
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
	cloudStorage file_db.CloudStorageClient,
) *Manager {
	return &Manager{
		store:        keyStore,
		filesystem:   filesyetem,
		cloudStorage: cloudStorage,
		config:       config,
	}
}
