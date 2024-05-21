package core

import (
	"bytes"
	"crypto/rand"
	"errors"

	"golang.org/x/crypto/chacha20poly1305"
)

// Key represents a 256-bit key used for ChaCha20-Poly1305.
type Key struct {
	value []byte
}

// NewKeyFromRand returns a new key generated from random bytes
func NewKeyFromRand() Key {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}
	return Key{
		value: key,
	}
}

// KeyFromBytes returns a new key from a byte slice
func KeyFromBytes(bytes []byte) (Key, error) {
	// Check if the slice length is exactly 3
	if len(bytes) != chacha20poly1305.KeySize {
		return EmptyKey(), errors.New("bytes does not contain exactly required bytes")
	}
	return Key{
		value: bytes,
	}, nil
}

// EmptyKey returns an empty key
func EmptyKey() Key {
	return Key{}
}

// Bytes returns the key as a byte slice
func (k *Key) Bytes() []byte {
	return k.value[:]
}

// Equals returns true if the key is equal to the other key
func (k *Key) Equals(other Key) bool {
	return bytes.Equal(k.value[:], other.value[:])
}

// IsEmpty returns true if the key is empty
func (k *Key) IsEmpty() bool {
	if k.Equals(Key{}) {
		return true
	}
	// Check if all bytes are 0
	for _, b := range k.value {
		if b != 0 {
			return false
		}
	}
	return true
}

// KeyInfo represents a key and its id
type KeyInfo struct {
	Key Key
	Id  string
}

// NewKeyInfo returns a new KeyInfo struct with the given key and id
func NewKeyInfo(keyId string, key Key) KeyInfo {
	return KeyInfo{
		Key: key,
		Id:  keyId,
	}
}
