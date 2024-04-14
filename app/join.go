package app

import (
	"ctb-cli/core"
)

// Join joins the user in the repository by creating the user key folder.
// Returns an AppResult indicating the success or failure of the operation.
func (a *App) Join(encryptedPrivateKey string) core.AppResult {
	// init the app
	initRes := a.initServices()
	if !initRes.Ok {
		return initRes
	}
	// set the private key
	setResult := a.SetPrivateKey(encryptedPrivateKey)
	if !setResult.Ok {
		return setResult
	}
	// Join the user using the key store
	err := a.keyStore.Join()
	if err != nil {
		return core.NewAppResultWithError(err)
	}
	return core.NewAppResult()
}

// JoinByUserId joins the user identified by the given userId.
// It initializes the app services, then joins the user using the key store.
// Returns an AppResult indicating the success or failure of the operation.
func (a *App) JoinByUserId(userId string) core.AppResult {
	// init the app
	initRes := a.initServices()
	if !initRes.Ok {
		return initRes
	}
	// Join the user using the key store
	err := a.keyStore.JoinByUserId(userId)
	if err != nil {
		return core.NewAppResultWithError(err)
	}
	return core.NewAppResult()
}
