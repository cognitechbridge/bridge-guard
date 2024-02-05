package manager

import (
	"ctb-cli/file_db"
)

type Manager struct {
	cloudStorage file_db.CloudStorageClient
}

var Client = Manager{}

func (mn *Manager) Init(
	cloudStorage file_db.CloudStorageClient,
) {
	mn.cloudStorage = cloudStorage
}
