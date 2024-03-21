package core

import (
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

// Key represents a 256-bit key used for ChaCha20-Poly1305.
type Key [chacha20poly1305.KeySize]byte

// NewKeyFromRand returns a new key generated from random bytes
func NewKeyFromRand() *Key {
	key := Key{}
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		panic(err)
	}
	return &key
}

// KeyFromBytes returns a new key from a byte slice
func KeyFromBytes(bytes []byte) (Key, error) {
	var key = Key{}

	// Check if the slice length is exactly 3
	if len(bytes) != chacha20poly1305.KeySize {
		return Key{}, errors.New("bytes does not contain exactly required bytes")
	}

	// Copy elements from the slice to the array
	copy(key[:], bytes)

	return key, nil
}

// KeyInfo represents a key and its id
type KeyInfo struct {
	Key []byte
	Id  string
}

// NewKeyInfo returns a new KeyInfo struct with the given key and id
func NewKeyInfo(keyId string, key []byte) KeyInfo {
	return KeyInfo{
		Key: key,
		Id:  keyId,
	}
}
