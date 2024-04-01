package app

import "ctb-cli/core"

// GenerateUserKey generates a user private key and returns it as a string.
// It returns an AppResult containing the generated key on success,
// or an AppErrorResult containing the error on failure.
func (a *App) GenerateUserKey() core.AppResult {
	// init the app to be able to use the key store
	initRes := a.initServices()
	if !initRes.Ok {
		return initRes
	}
	// generate the key
	key, err := a.keyStore.GenerateUserKey()
	if err != nil {
		return core.NewAppResultWithError(err)
	}
	return core.NewAppResultWithValue(*key)
}
