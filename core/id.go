package core

import (
	"crypto/rand"
	"errors"
	"io"

	"github.com/btcsuite/btcutil/base58"
)

var (
	ErrInvalidPublicKey = errors.New("invalid public key")
)

// KeyService represents the service used to manage keys
func encodedPrivateKey(privateKey []byte) string {
	return base58.Encode(privateKey)
}

// NewKeyFromRand returns a new key generated from random bytes
func EncodePublic(byte []byte) string {
	return base58.Encode(byte)
}

// DecodePublic decodes a public key from a string using base58 encoding
func DecodePublic(str string) ([]byte, error) {
	pub := base58.Decode(str)
	return pub, nil
}

// DecodePrivateKey decodes a private key from a string using base58 encoding
func DecodePrivateKey(str string) ([]byte, error) {
	return base58.Decode(str), nil
}

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
