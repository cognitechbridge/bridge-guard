package app

import "ctb-cli/core"

// GetStatus returns the status of the repository.
// It checks if the repository is valid and if the user has joined.
// Returns an AppResult with the repository status.
func GetStatus() core.AppResult {
	valid, err := fileSystem.IsValidRepository()
	if err != nil {
		return core.AppErrorResult(err)
	}
	isJoined := keyStore.IsUserJoined()
	return core.AppOkResultWithResult(core.RepositroyStatus{
		IsValid:  valid,
		IsJoined: isJoined,
	})
}
