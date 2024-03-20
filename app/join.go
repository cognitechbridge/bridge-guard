package app

import (
	"ctb-cli/core"
)

// Join joins the user in the repository by creating the user key folder.
// Returns an AppResult indicating the success or failure of the operation.
func Join(encryptedPrivateKey string) core.AppResult {
	// init the app
	initRes := initApp()
	if !initRes.Ok {
		return initRes
	}
	// set the private key
	setResult := SetPrivateKey(encryptedPrivateKey)
	if !setResult.Ok {
		return setResult
	}
	err := keyStore.Join()
	if err != nil {
		return core.AppErrorResult(err)
	}
	return core.AppOkResult()
}
