package app

import "ctb-cli/core"

func SetAndCheckSecret(secret string) core.AppResult {
	keyStore.SetSecret(secret)
	if err := keyStore.LoadKeys(); err != nil {
		return core.AppErrorResult(err)
	}
	return core.AppOkResult()
}
