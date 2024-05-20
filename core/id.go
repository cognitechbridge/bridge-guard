package core

import (
	"crypto/rand"
	"io"

	"github.com/btcsuite/btcutil/base58"
)

// NewUid returns a new uid as a string
func NewUid() (string, error) {
	rnd := make([]byte, 32)
	_, _ = io.ReadFull(rand.Reader, rnd)
	return EncodeUid(rnd)
}

// EncodeUid encodes a uid to a string using base58 encoding
func EncodeUid(uid []byte) (string, error) {
	return base58.Encode(uid), nil
}
