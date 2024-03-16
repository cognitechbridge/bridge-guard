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
	"errors"
	"os"
	"path/filepath"
)

var fileSystem *filesyetem_service.FileSystem
var keyStore core.KeyService
var shareService *share_service.Service

var (
	ErrPrivateKeyCheckFailed = errors.New("private key check failed")
)

func Init() {
	cloudClient := cloud.NewClient("http://localhost:1323", 10*1024*1024)
	//cloudClient := objectstorage.NewDummyClient()

	userId, _ := config.Workspace.GetUserId()

	root, _ := config.GetRepoCtbRoot()
	tempRoot, _ := config.GetTempRoot()

	keysPath := createAndReturn(filepath.Join(root, "keys"))
	objectPath := createAndReturn(filepath.Join(root, "object"))
	recipientsPath := createAndReturn(filepath.Join(root, "recipients"))
	filesystemPath := createAndReturn(filepath.Join(root, "filesystem"))
	cachePath := createAndReturn(filepath.Join(tempRoot, "cache"))
	vaultPath := createAndReturn(filepath.Join(root, "vault"))

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

func createAndReturn(path string) string {
	os.MkdirAll(path, os.ModePerm)
	return path
}

func SetAndCheckPrivateKey(encodedPrivateKey string) core.AppResult {
	privateKey, err := core.DecodePrivateKey(encodedPrivateKey)
	if err != nil {
		return core.AppErrorResult(err)
	}
	keyStore.SetPrivateKey(privateKey)
	res, err := keyStore.CheckPrivateKey()
	if !res {
		return core.AppErrorResult(ErrPrivateKeyCheckFailed)
	}
	return core.AppOkResult()
}
