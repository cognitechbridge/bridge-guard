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
}

func NewManager(store *keystore.KeyStore, filesystem *filesyetem.FileSystem, cloudStorage persist_file.CloudStorageClient) *Manager {
	return &Manager{
		store:        store,
		filesystem:   filesystem,
		cloudStorage: cloudStorage,
	}
}
