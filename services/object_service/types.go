package object_service

import "ctb-cli/core"

// encryptChanItem represents an item to be encrypted.
type encryptChanItem struct {
	id  string
	dir string
	key *core.KeyInfo
}

// uploadChanItem represents an item to be uploaded.
type uploadChanItem struct {
	id  string
	dir string
}
