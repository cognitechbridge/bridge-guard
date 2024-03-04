package app

import "ctb-cli/fuse"

func Mount() {
	ctbFuse := fuse.New(fileSystem)
	ctbFuse.Mount()
}
