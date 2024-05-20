package core

import (
	"crypto/rand"

	"golang.org/x/crypto/curve25519"
)

type PublicKey []byte

// NewPublicKeyFromEncoded creates a PublicKey from an encoded base58 string.
func NewPublicKeyFromEncoded(encoded string) (PublicKey, error) {
	if len(encoded) != 44 {
		return nil, ErrInvalidPublicKey
	}
	return DecodePublic(encoded)
}

// NewPublicKeyFromBytes creates a PublicKey from a byte slice.
func NewPublicKeyFromBytes(bytes []byte) PublicKey {
	return PublicKey(bytes)
}

// String returns the base58 encoded string representation of the PublicKey.
func (key PublicKey) String() string {
	return key.Encode()
}

// Encode returns the base58 encoded string representation of the PublicKey.
func (key PublicKey) Encode() string {
	return EncodePublic(key)
}

// Bytes returns the byte slice representation of the PublicKey.
func (key PublicKey) Bytes() []byte {
	return []byte(key)
}

// Equals returns true if the PublicKey is equal to the other PublicKey.
func (key PublicKey) Equals(other PublicKey) bool {
	return key.String() == other.String()
}

type PrivateKey []byte

// NewPrivateKeyFromEncoded creates a PrivateKey from an encoded base58 string.
func NewPrivateKeyFromEncoded(encoded string) (PrivateKey, error) {
	return DecodePrivateKey(encoded)
}

// NewPrivateKeyFromBytes creates a PrivateKey from a byte slice.
func NewPrivateKeyFromBytes(bytes []byte) PrivateKey {
	return PrivateKey(bytes)
}

// String returns the base58 encoded string representation of the PrivateKey.
func (key PrivateKey) String() string {
	return key.Encode()
}

// Encode returns the base58 encoded string representation of the PrivateKey.
func (key PrivateKey) Encode() string {
	return encodedPrivateKey(key)
}

// Bytes returns the byte slice representation of the PrivateKey.
func (key PrivateKey) Bytes() []byte {
	return []byte(key)
}

// Equals returns true if the PrivateKey is equal to the other PrivateKey.
func (key PrivateKey) Equals(other PrivateKey) bool {
	return key.String() == other.String()
}

type UserKeyPair struct {
	PublicKey  string
	PrivateKey string
}

func GenerateKey() (*KeyInfo, error) {
	key := NewKeyFromRand()
	keyId, err := NewUid()
	if err != nil {
		return nil, err
	}
	keyInfo := NewKeyInfo(keyId, key[:])
	return &keyInfo, nil
}

func GenerateUserKey() (*UserKeyPair, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	encodedPrivateKey := encodedPrivateKey(key)
	publicKey, err := curve25519.X25519(key, curve25519.Basepoint)
	if err != nil {
		return nil, err
	}
	encodedPublicKey := EncodePublic(publicKey)
	return &UserKeyPair{
		PublicKey:  encodedPublicKey,
		PrivateKey: encodedPrivateKey,
	}, nil
}
