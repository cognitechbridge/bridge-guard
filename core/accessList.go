package core

type KeyAccess struct {
	PublicKey string
	Inherited bool
}

type KeyAccessList = []KeyAccess
