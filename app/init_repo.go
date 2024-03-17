package app

import "ctb-cli/core"

// InitRepo initializes the repository.
// It creates a vault in the root path.
// Returns an AppResult indicating the success or failure of the operation.
func InitRepo() core.AppResult {
	// Create a vault in the root path
	if err := fileSystem.CreateVaultInPath("/"); err != nil {
		core.AppErrorResult(err)
	}
	return core.AppOkResult()
}
