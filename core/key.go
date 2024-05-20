package core

import (
	"crypto/rand"

	"golang.org/x/crypto/curve25519"
)

type PublicKey struct {
	value []byte
}

func EmptyPublicKey() PublicKey {
	return PublicKey{}
}

// NewPublicKeyFromEncoded creates a PublicKey from an encoded base58 string.
func NewPublicKeyFromEncoded(encoded string) (PublicKey, error) {
	if len(encoded) != 44 {
		return EmptyPublicKey(), ErrInvalidPublicKey
	}
	return PublicKey{
		value: DecodePublic(encoded),
	}, nil
}

// NewPublicKeyFromBytes creates a PublicKey from a byte slice.
func NewPublicKeyFromBytes(bytes []byte) PublicKey {
	return PublicKey{
		value: bytes,
	}
}

// String returns the base58 encoded string representation of the PublicKey.
func (key PublicKey) String() string {
	return key.Encode()
}

// MarshalJSON returns the JSON encoding of the PublicKey.
func (key PublicKey) MarshalJSON() ([]byte, error) {
	return []byte(`"` + key.Encode() + `"`), nil
}

// Encode returns the base58 encoded string representation of the PublicKey.
func (key PublicKey) Encode() string {
	return EncodePublic(key.value)
}

// Bytes returns the byte slice representation of the PublicKey.
func (key PublicKey) Bytes() []byte {
	return key.value
}

// Equals returns true if the PublicKey is equal to the other PublicKey.
func (key PublicKey) Equals(other PublicKey) bool {
	return key.String() == other.String()
}

type PrivateKey struct {
	value []byte
}

func EmptyPrivateKey() PrivateKey {
	return PrivateKey{}
}

// NewPrivateKeyFromEncoded creates a PrivateKey from an encoded base58 string.
func NewPrivateKeyFromEncoded(encoded string) (PrivateKey, error) {
	if len(encoded) != 44 {
		return EmptyPrivateKey(), ErrInvalidPublicKey
	}
	return PrivateKey{
		value: DecodePrivateKey(encoded),
	}, nil
}

// NewPrivateKeyFromBytes creates a PrivateKey from a byte slice.
func NewPrivateKeyFromBytes(bytes []byte) PrivateKey {
	return PrivateKey{
		value: bytes,
	}
}

// Bytes returns the byte slice representation of the PrivateKey.
func (key PrivateKey) Bytes() []byte {
	return key.value
}

// Equals returns true if the PrivateKey is equal to the other PrivateKey.
func (key PrivateKey) Equals(other PrivateKey) bool {
	for i := range key.value {
		if key.value[i] != other.value[i] {
			return false
		}
	}
	return true
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
