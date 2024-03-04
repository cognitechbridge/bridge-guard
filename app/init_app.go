package app

import (
	"ctb-cli/config"
	"ctb-cli/core"
	"ctb-cli/objectstorage/cloud"
	"ctb-cli/prompts"
	"ctb-cli/repositories"
	"ctb-cli/services/filesyetem_service"
	"ctb-cli/services/key_service"
	"ctb-cli/services/object_service"
	"ctb-cli/services/share_service"
	"errors"
	"github.com/fatih/color"
	"os"
	"path/filepath"
)

var fileSystem *filesyetem_service.FileSystem
var keyStore core.KeyService
var shareService *share_service.Service

func InitSecret(secret string) error {
	needSecret := secret == "" // Determine if we need to prompt for the secret

	for {
		if needSecret {
			var err error
			secret, err = prompts.GetSecret()
			if err != nil {
				return err // If there's an error getting the secret, return immediately
			}
		}
		keyStore.SetSecret(secret)
		err := keyStore.LoadKeys()
		if err == nil {
			return nil // Success, exit function
		}

		if errors.Is(err, key_service.ErrorInvalidSecret) {
			// Notify user of invalid secret
			c := color.New(color.FgRed, color.Bold)
			_, _ = c.Println("Invalid secret. Try again")

			if !needSecret {
				return err
			}
		} else {
			return err // For any other error, return immediately
		}
	}
}

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
