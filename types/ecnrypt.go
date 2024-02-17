package types

import (
	"crypto/rand"
	"crypto/rsa"
	"golang.org/x/crypto/chacha20poly1305"
	"io"
)

// Key represents a 256-bit key used for ChaCha20-Poly1305.
type Key [chacha20poly1305.KeySize]byte

func NewKeyFromRand() *Key {
	key := Key{}
	io.ReadFull(rand.Reader, key[:])
	return &key
}

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
