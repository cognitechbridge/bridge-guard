package app

import "ctb-cli/core"

func ChangeSecret(secret string) core.AppResult {
	if err := keyStore.ChangeSecret(secret); err != nil {
		return core.AppErrorResult(err)
	}
	return core.AppOkResult()
}
