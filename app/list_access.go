package app

import "ctb-cli/core"

func (a *App) ListAccess(path string) core.AppResult {
	// init the app
	initRes := a.initServices()
	if !initRes.Ok {
		return initRes
	}
	res, err := a.shareService.GetAccessList(path)
	if err != nil {
		return core.NewAppResultWithError(err)
	}
	return core.NewAppResultWithValue(res)
}
