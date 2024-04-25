package app

import "ctb-cli/core"

// GetStatus returns the status of the repository.
// It checks if the repository is valid and if the user has joined.
// Returns an AppResult with the repository status.
func (a *App) GetStatus(encryptedPrivateKey string) core.AppResult {
	// check if the root folder is empty
	empty := a.IsRootEmpty()
	if empty {
		return core.NewAppResultWithValue(core.NewInvalidRepositoyStatus(true))
	}

	// check if the repository is valid
	valid := a.configService.IsRepositoryConfigExists("")

	if !valid {
		// return the repository status with IsValid = false if the repository config does not exist
		return core.NewAppResultWithValue(core.NewInvalidRepositoyStatus(false))
	}

	// init the app
	initRes := a.initServices()
	if !initRes.Ok {
		// return the repository status with IsValid = false if the app failed to initialize the services
		return core.NewAppResultWithValue(core.NewInvalidRepositoyStatus(false))
	}

	// set the private key if it was passed
	if encryptedPrivateKey != "" {
		a.SetAndCheckPrivateKey(encryptedPrivateKey)
	}

	// check if the user has joined
	isJoined := a.keyStore.IsUserJoined()
	publicKey := ""
	if isJoined {
		var err error
		if publicKey, err = a.keyStore.GetEncodablePublicKey(); err != nil {
			return core.NewAppResultWithError(err)
		}
	}
	return core.NewAppResultWithValue(core.RepositoryStatus{
		IsValid:   valid,
		IsJoined:  isJoined,
		RepoId:    "",
		PublicKey: publicKey,
	})
}
