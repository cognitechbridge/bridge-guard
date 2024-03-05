package app

import "ctb-cli/core"

func InitRepo() core.AppResult {
	if err := fileSystem.CreateVaultInPath("/"); err != nil {
		core.AppErrorResult(err)
	}
	return core.AppOkResult()
}
