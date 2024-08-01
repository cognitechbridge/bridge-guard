package app

import (
	"ctb-cli/core"
	"ctb-cli/fuse"
)

// Mount mounts the file system and returns the result.
// It returns an AppResult containing the result of the operation.
func (a *App) Mount() core.AppResult {
	a.fuse.Mount()
	return core.NewAppResult()
}

// PrepareMount creates the fuse file system and returns the result.
func (a *App) PrepareMount(encryptedPrivateKey string, mount string) core.AppResult {
	// init the app
	initRes := a.initServices()
	if !initRes.Ok {
		return initRes
	}
	// set the private key
	keySetRes := a.SetAndCheckPrivateKey(encryptedPrivateKey)
	if !keySetRes.Ok {
		return keySetRes
	}
	// create the fuse
	a.fuse = fuse.New(a.fileSystem)
	res := a.fuse.FindMountPoint(mount)
	return core.NewAppResultWithValue(res)
}
