package types

import (
	"crypto/rsa"
	"golang.org/x/crypto/chacha20poly1305"
)

// Key represents a 256-bit key used for ChaCha20-Poly1305.
type Key [chacha20poly1305.KeySize]byte

type KeyInfo struct {
	Key           []byte
	Id            string
	RecoveryBlobs []string
}

type SerializedKey struct {
	ID    string
	Nonce string
	Key   string
}

type RecoveryItem struct {
	PublicKey *rsa.PublicKey
	Sha1      string
}
