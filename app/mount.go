package app

import (
	"ctb-cli/core"
	"ctb-cli/fuse"
)

// Mount mounts the file system and returns the result.
// It returns an AppResult containing the result of the operation.
func Mount(encryptedPrivateKey string) core.AppResult {
	// init the app
	initRes := initApp()
	if !initRes.Ok {
		return initRes
	}
	// set the private key
	keySetRes := SetAndCheckPrivateKey(encryptedPrivateKey)
	if !keySetRes.Ok {
		return keySetRes
	}
	ctbFuse := fuse.New(fileSystem)
	ctbFuse.Mount()
	return core.AppOkResult()
}
