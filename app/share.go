package app

import "ctb-cli/core"

// Share shares a file or directory located at the specified path with the given public key.
// Returns an AppResult indicating the success or failure of the operation.
func (a *App) Share(path string, publicKey string, encryptedPrivateKey string) core.AppResult {
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
	if err := a.shareService.ShareByPublicKey(path, publicKey); err != nil {
		return core.NewAppResultWithError(err)
	}
	return core.NewAppResult()
}

// Unshare removes the sharing of a file or directory with a specific public key.
// It initializes the app services and calls the UnshareByPublicKey method of the shareService.
// If an error occurs during the unsharing process, it returns an AppResult with the error.
// Otherwise, it returns a successful AppResult.
func (a *App) Unshare(path string, publicKey string) core.AppResult {
	// init the app
	initRes := a.initServices()
	if !initRes.Ok {
		return initRes
	}
	if err := a.shareService.Unshare(path, publicKey); err != nil {
		return core.NewAppResultWithError(err)
	}
	return core.NewAppResult()
}
