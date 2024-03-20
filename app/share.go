package app

import "ctb-cli/core"

// Share shares the files that match the pattern with the given public key.
// Returns an AppResult indicating the success or failure of the operation.
func Share(pattern string, publicKey string, encryptedPrivateKey string) core.AppResult {
	// init the app
	initRes := initApp()
	if !initRes.Ok {
		return initRes
	}
	// set the private key
	keySetRes := SetAndCheckPrivateKey(encryptedPrivateKey)
	if !keySetRes.Ok {
		return keySetRes
	}
	if err := shareService.ShareByPublicKey(pattern, publicKey); err != nil {
		return core.AppErrorResult(err)
	}
	return core.AppOkResult()
}
