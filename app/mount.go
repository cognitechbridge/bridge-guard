package app

import (
	"ctb-cli/core"
	"ctb-cli/fuse"
)

// Mount mounts the file system and returns the result.
// It returns an AppResult containing the result of the operation.
func Mount() core.AppResult {
	ctbFuse := fuse.New(fileSystem)
	ctbFuse.Mount()
	return core.AppOkResult()
}
