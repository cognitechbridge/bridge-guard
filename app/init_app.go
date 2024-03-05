package app

import (
	"ctb-cli/config"
	"ctb-cli/core"
	"ctb-cli/objectstorage/cloud"
	"ctb-cli/repositories"
	"ctb-cli/services/filesyetem_service"
	"ctb-cli/services/key_service"
	"ctb-cli/services/object_service"
	"ctb-cli/services/share_service"
	"os"
	"path/filepath"
)

var fileSystem *filesyetem_service.FileSystem
var keyStore core.KeyService
var shareService *share_service.Service

func Init() {
	cloudClient := cloud.NewClient("http://localhost:1323", 10*1024*1024)
	//cloudClient := objectstorage.NewDummyClient()

	userId, _ := config.Workspace.GetUserId()

	root, _ := config.GetRepoCtbRoot()
	tempRoot, _ := config.GetTempRoot()

	keysPath := CreateAndReturn(filepath.Join(root, "keys"))
	objectPath := CreateAndReturn(filepath.Join(root, "object"))
	recipientsPath := CreateAndReturn(filepath.Join(root, "recipients"))
	filesystemPath := CreateAndReturn(filepath.Join(root, "filesystem"))
	cachePath := CreateAndReturn(filepath.Join(tempRoot, "cache"))
	vaultPath := CreateAndReturn(filepath.Join(root, "vault"))

	keyRepository := repositories.NewKeyRepositoryFile(keysPath)
	objectCacheRepository := repositories.NewObjectCacheRepository(cachePath)
	objectRepository := repositories.NewObjectRepository(objectPath)
	recipientRepository := repositories.NewRecipientRepositoryFile(recipientsPath)
	linkRepository := repositories.NewLinkRepository(filesystemPath)
	vaultRepository := repositories.NewVaultRepositoryFile(vaultPath)

	keyStore = key_service.NewKeyStore(userId, keyRepository, vaultRepository)

	objectService := object_service.NewService(userId, &objectCacheRepository, &objectRepository, cloudClient)
	shareService = share_service.NewService(recipientRepository, keyStore, linkRepository, &objectService)

	fileSystem = filesyetem_service.NewFileSystem(keyStore, objectService, linkRepository)
}

func CreateAndReturn(path string) string {
	os.MkdirAll(path, os.ModePerm)
	return path
}
