package app

import "ctb-cli/core"

func Join() core.AppResult {
	keyStore.Join()
	return core.AppOkResult()
}
