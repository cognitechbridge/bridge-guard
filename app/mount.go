package app

import (
	"ctb-cli/core"
	"ctb-cli/fuse"
)

func Mount() core.AppResult {
	ctbFuse := fuse.New(fileSystem)
	ctbFuse.Mount()
	return core.AppOkResult()
}
