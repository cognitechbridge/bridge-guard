package manager

import (
	"ctb-cli/file_db"
	"ctb-cli/filesyetem"
	"ctb-cli/keystore"
)

type Manager struct {
	store        *keystore.KeyStore
	cloudStorage file_db.CloudStorageClient
	Filesystem   *filesyetem.FileSystem
}

var Client = Manager{}

func (mn *Manager) Init(
	keyStore *keystore.KeyStore,
	filesyetem *filesyetem.FileSystem,
	cloudStorage file_db.CloudStorageClient,
) {
	mn.cloudStorage = cloudStorage
	mn.Filesystem = filesyetem
	mn.store = keyStore
}
