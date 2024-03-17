package app

import "ctb-cli/core"

func GenerateUserKey() core.AppResult {
	key, err := keyStore.GenerateUserKey()
	if err != nil {
		return core.AppErrorResult(err)
	}
	return core.AppOkResultWithResult(key)
}
