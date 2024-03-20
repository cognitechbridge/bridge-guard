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
	ErrPrivateKeyCheckFailed     = errors.New("private key check failed")
	ErrInvalidPrivateKeySize     = errors.New("invalid private key size")
	ErrCreatingRepositoryFolders = errors.New("error creating repository folders")
	ErrRootFolderNotEmpty        = errors.New("root folder is not empty")
)

func initApp() core.AppResult {
	cloudClient := cloud.NewClient("http://localhost:1323", 10*1024*1024)
	//cloudClient := objectstorage.NewDummyClient()

	// Get the root paths
	root, _ := config.GetRepoCtbRoot()
	tempRoot, _ := config.GetTempRoot()

	// Create the repository paths
	keysPath := filepath.Join(root, "keys")
	objectPath := filepath.Join(root, "object")
	filesystemPath := filepath.Join(root, "filesystem")
	cachePath := filepath.Join(tempRoot, "cache")
	vaultPath := filepath.Join(root, "vault")

	// Check if the paths exist
	err := errors.Join(
		checkFolderPath(keysPath),
		checkFolderPath(objectPath),
		checkFolderPath(filesystemPath),
		checkFolderPath(cachePath),
		checkFolderPath(vaultPath),
	)

	// If at least one path doesn't exist, panic
	if err != nil {
		return core.AppErrorResult(err)
	}

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

	return core.AppOkResult()
}

// checkFolderPath checks if the path exists and returns an error if it doesn't.
func checkFolderPath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	return nil
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

// InitRepo initializes the repository by creating the necessary folders, setting the private key,
// and joining the user. It also creates a vault in the root path.
// The encryptedPrivateKey parameter is the encrypted private key used for authentication.
// It returns an AppResult indicating the success or failure of the initialization.
func InitRepo(encryptedPrivateKey string) core.AppResult {
	// Get the root and temp paths
	root, _ := config.GetRepoCtbRoot()
	tempRoot, _ := config.GetTempRoot()

	// Check if the root folder is empty
	rootFiles, err := os.ReadDir(root)
	if err != nil {
		return core.AppErrorResult(err)
	}
	if len(rootFiles) > 0 {
		return core.AppErrorResult(ErrRootFolderNotEmpty)
	}

	// Create the repository folders
	err = errors.Join(
		os.MkdirAll(filepath.Join(root, "keys"), os.ModePerm),
		os.MkdirAll(filepath.Join(root, "filesystem"), os.ModePerm),
		os.MkdirAll(filepath.Join(root, "object"), os.ModePerm),
		os.MkdirAll(filepath.Join(tempRoot, "cache"), os.ModePerm),
		os.MkdirAll(filepath.Join(root, "vault"), os.ModePerm),
	)
	if err != nil {
		return core.AppErrorResult(ErrCreatingRepositoryFolders)
	}

	// init the app
	initRes := initApp()
	if !initRes.Ok {
		return initRes
	}

	// Join the user
	joinResult := Join(encryptedPrivateKey)
	if !joinResult.Ok {
		return joinResult
	}

	// Create a vault in the root path
	if err := fileSystem.CreateVaultInPath("/"); err != nil {
		return core.AppErrorResult(err)
	}
	return core.AppOkResult()
}
