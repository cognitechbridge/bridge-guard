package secure_storage

import (
	"storage-go/filesyetem"
	"storage-go/keystore"
	"storage-go/storage"
)

type Manager struct {
	store        *keystore.KeyStore
	s3storage    *storage.S3Storage
	cloudStorage storage.CloudStorageClient
	filesystem   *filesyetem.FileSystem
}

func NewManager(store *keystore.KeyStore, s3storage *storage.S3Storage, filesystem *filesyetem.FileSystem, cloudStorage storage.CloudStorageClient) *Manager {
	return &Manager{
		store:        store,
		s3storage:    s3storage,
		filesystem:   filesystem,
		cloudStorage: cloudStorage,
	}
}
