package app

import (
	"ctb-cli/core"
)

// @Todo: Remove!

// Join joins the user in the repository by creating the user key folder.
// Returns an AppResult indicating the success or failure of the operation.
func (a *App) Join(encryptedPrivateKey string) core.AppResult {
	return core.NewAppResult()
}

// JoinByUserId joins the user identified by the given userId.
// It initializes the app services, then joins the user using the key store.
// Returns an AppResult indicating the success or failure of the operation.
func (a *App) JoinByUserId(userId string) core.AppResult {
	return core.NewAppResult()
}
