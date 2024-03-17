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
	ErrInvalidPrivateKeySize = errors.New("invalid private key size")
)

func Init() {
	cloudClient := cloud.NewClient("http://localhost:1323", 10*1024*1024)
	//cloudClient := objectstorage.NewDummyClient()

	// Get the root paths
	root, _ := config.GetRepoCtbRoot()
	tempRoot, _ := config.GetTempRoot()

	// Get the repository paths and create them if they don't exist
	keysPath := createAndReturn(filepath.Join(root, "keys"))
	objectPath := createAndReturn(filepath.Join(root, "object"))
	filesystemPath := createAndReturn(filepath.Join(root, "filesystem"))
	cachePath := createAndReturn(filepath.Join(tempRoot, "cache"))
	vaultPath := createAndReturn(filepath.Join(root, "vault"))

	// Create the repositories
	keyRepository := repositories.NewKeyRepositoryFile(keysPath)
	objectCacheRepository := repositories.NewObjectCacheRepository(cachePath)
	objectRepository := repositories.NewObjectRepository(objectPath)
	linkRepository := repositories.NewLinkRepository(filesystemPath)
	vaultRepository := repositories.NewVaultRepositoryFile(vaultPath)

	// Create the services
	keyStore = key_service.NewKeyStore(keyRepository, vaultRepository)
	objectService := object_service.NewService(&objectCacheRepository, &objectRepository, cloudClient)
	shareService = share_service.NewService(keyStore, linkRepository, &objectService)
	fileSystem = filesyetem_service.NewFileSystem(keyStore, objectService, linkRepository)
}

// CreateAndReturn creates a directory and returns the path.
func createAndReturn(path string) string {
	os.MkdirAll(path, os.ModePerm)
	return path
}

// SetPrivateKey sets the private key used by the application.
// It takes an encoded private key as input and returns an AppResult.
// If the private key is successfully decoded and its size is valid, it is set in the keyStore.
// Otherwise, an error result is returned.
func SetPrivateKey(encodedPrivateKey string) core.AppResult {
	// Decode the private key
	privateKey, err := core.DecodePrivateKey(encodedPrivateKey)
	if err != nil {
		return core.AppErrorResult(err)
	}
	// Check the size of the private key
	if len(privateKey) != 32 {
		return core.AppErrorResult(ErrInvalidPrivateKeySize)
	}
	// Set the private key in the keyStore
	keyStore.SetPrivateKey(privateKey)
	return core.AppOkResult()
}

// SetAndCheckPrivateKey sets the private key and checks its validity.
// It takes an encodedPrivateKey as input and returns an AppResult indicating the success or failure of the operation.
func SetAndCheckPrivateKey(encodedPrivateKey string) core.AppResult {
	// Set the private key
	setResult := SetPrivateKey(encodedPrivateKey)
	if !setResult.Ok {
		return setResult
	}
	// Check the private key
	res, _ := keyStore.CheckPrivateKey()
	if !res {
		return core.AppErrorResult(ErrPrivateKeyCheckFailed)
	}
	return core.AppOkResult()
}
