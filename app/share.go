package app

import "ctb-cli/core"

func Share(pattern string, recipient string) core.AppResult {
	if err := shareService.ShareByPublicKey(pattern, recipient); err != nil {
		return core.AppErrorResult(err)
	}
	return core.AppOkResult()
}
