package object_service

import "ctb-cli/core"

type encryptChanItem struct {
	id  string
	key *core.KeyInfo
}

type uploadChanItem struct {
	id string
}
