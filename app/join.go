package app

import "ctb-cli/core"

// Join joins the user in the repository by creating the user key folder.
// Returns an AppResult indicating the success or failure of the operation.
func Join() core.AppResult {
	err := keyStore.Join()
	if err != nil {
		return core.AppErrorResult(err)
	}
	return core.AppOkResult()
}
