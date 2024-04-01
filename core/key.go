package core

import (
	"crypto/rand"

	"golang.org/x/crypto/curve25519"
)

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
	encodedPrivateKey, err := encodedPrivateKey(key)
	if err != nil {
		return nil, err
	}
	publicKey, err := curve25519.X25519(key, curve25519.Basepoint)
	if err != nil {
		return nil, err
	}
	encodedPublicKey, err := EncodePublic(publicKey)
	if err != nil {
		return nil, err
	}
	return &UserKeyPair{
		PublicKey:  encodedPublicKey,
		PrivateKey: encodedPrivateKey,
	}, nil
}
