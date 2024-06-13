package app

import (
	"ctb-cli/config"
	"ctb-cli/core"
	"ctb-cli/fuse"
	"ctb-cli/objectstorage/cloud"
	"ctb-cli/repositories"
	"ctb-cli/services/config_service"
	"ctb-cli/services/filesystem_service"
	"ctb-cli/services/key_service"
	"ctb-cli/services/object_service"
	"ctb-cli/services/share_service"
	"errors"
	"os"
	"path/filepath"
)

// App represents the main application struct.
type App struct {
	keyStore      core.KeyService
	fileSystem    *filesystem_service.FileSystem
	shareService  *share_service.Service
	configService *config_service.ConfigService

	// fuse is the fuse service used by the application
	fuse *fuse.CtbFs

	// Config is the configuration of the application
	cfg *config.Config
}

var (
	ErrPrivateKeyCheckFailed     = errors.New("private key check failed")
	ErrInvalidPrivateKeySize     = errors.New("invalid private key size")
	ErrCreatingRepositoryFolders = errors.New("error creating repository folders")
	ErrRootFolderNotEmpty        = errors.New("root folder is not empty")
	ErrCreatingRepositoryConfig  = errors.New("error creating repository config")
	ErrInitRepositoryFolders     = errors.New("error initializing repository folders")
)

// New returns a new App
func New(cfg config.Config) App {
	return App{
		cfg: &cfg,
	}
}

func (a *App) initServices() core.AppResult {
	cloudClient := cloud.NewClient("http://localhost:1323", 10*1024*1024)
	//cloudClient := objectstorage.NewDummyClient()

	// Get the root paths
	root, _ := a.cfg.GetRepoCtbRoot()
	cachePath, _ := a.cfg.GetCacheRoot()

	// Create the repositories
	keyRepository := repositories.NewKeyRepositoryFile(root)
	objectCacheRepository := repositories.NewObjectCacheRepository(cachePath)
	objectRepository := repositories.NewObjectRepository(root)
	linkRepository := repositories.NewLinkRepository(root)
	vaultRepository := repositories.NewVaultRepositoryFile(root)

	// Create the services
	a.keyStore = key_service.NewKeyStore(keyRepository, vaultRepository)
	objectService := object_service.NewService(&objectCacheRepository, &objectRepository, cloudClient)
	a.shareService = share_service.NewService(a.keyStore, linkRepository, vaultRepository, &objectService)
	a.configService = config_service.New(root)
	a.fileSystem = filesystem_service.NewFileSystem(a.keyStore, objectService, linkRepository, vaultRepository, *a.configService)

	return core.NewAppResult()
}

// SetPrivateKey sets the private key used by the application.
// It takes an encoded private key as input and returns an AppResult.
// If the private key is successfully decoded and its size is valid, it is set in the keyStore.
// Otherwise, an error result is returned.
func (a *App) SetPrivateKey(encodedPrivateKey string) core.AppResult {
	// Decode the private key
	privateKey, err := core.NewPrivateKeyFromEncoded(encodedPrivateKey)
	if err != nil {
		return core.NewAppResultWithError(err)
	}
	// Set the private key in the keyStore
	a.keyStore.SetPrivateKey(privateKey)
	return core.NewAppResult()
}

// SetAndCheckPrivateKey sets the private key and checks its validity.
// It takes an encodedPrivateKey as input and returns an AppResult indicating the success or failure of the operation.
func (a *App) SetAndCheckPrivateKey(encodedPrivateKey string) core.AppResult {
	// Set the private key
	setResult := a.SetPrivateKey(encodedPrivateKey)
	if !setResult.Ok {
		return setResult
	}
	// Check the private key
	res := a.keyStore.IsUserJoined()
	if !res {
		return core.NewAppResultWithError(ErrPrivateKeyCheckFailed)
	}
	return core.NewAppResult()
}

// InitRepo initializes the repository by creating the necessary folders, setting the private key,
// and joining the user. It also creates a vault in the root path.
// The encryptedPrivateKey parameter is the encrypted private key used for authentication.
// It returns an AppResult indicating the success or failure of the initialization.
func (a *App) InitRepo(encryptedPrivateKey string) core.AppResult {
	// Get the root and temp paths
	root, _ := a.cfg.GetRepoCtbRoot()

	// Check if the root folder is empty
	rootFiles, err := os.ReadDir(root)
	if err != nil {
		return core.NewAppResultWithError(err)
	}
	if len(rootFiles) > 0 {
		return core.NewAppResultWithError(ErrRootFolderNotEmpty)
	}

	// Create the system folders
	systemFolders := core.GetRepoSystemFolderNames()
	for _, folder := range systemFolders {
		err := os.MkdirAll(filepath.Join(root, ".meta", folder), os.ModePerm)
		if err != nil {
			return core.NewAppResultWithError(ErrCreatingRepositoryFolders)
		}
	}

	// init the app
	initRes := a.initServices()
	if !initRes.Ok {
		return initRes
	}

	// Initialize the configuration
	err = a.configService.InitConfig("")
	if err != nil {
		return core.NewAppResultWithError(ErrCreatingRepositoryConfig)
	}

	// Set the private key
	setResult := a.SetPrivateKey(encryptedPrivateKey)
	if !setResult.Ok {
		return setResult
	}

	// Create a vault in the root path
	if err := a.fileSystem.CreateVaultInPath("/"); err != nil {
		return core.NewAppResultWithError(err)
	}
	return core.NewAppResult()
}

// IsRootEmpty checks if the root folder is empty.
// It returns true if the folder is empty, false otherwise.
func (a *App) IsRootEmpty() bool {
	root, _ := a.cfg.GetRepoCtbRoot()
	rootFiles, err := os.ReadDir(root)
	if err != nil {
		return false
	}
	return len(rootFiles) == 0
}
