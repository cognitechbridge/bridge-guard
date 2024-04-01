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
	valid := a.cfg.IsRepositoryConfigExists()

	if !valid {
		// return the repository status with IsValid = false if the repository config does not exist
		return core.NewAppResultWithValue(core.NewInvalidRepositoyStatus(false))
	}

	// get the repository id if it is valid
	repoId := a.cfg.GetRepoId()

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
	return core.NewAppResultWithValue(core.RepositoryStatus{
		IsValid:  valid,
		IsJoined: isJoined,
		RepoId:   repoId,
	})
}
