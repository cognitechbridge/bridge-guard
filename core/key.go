package core

import "crypto/rand"

func GenerateKey() (*KeyInfo, error) {
	key := NewKeyFromRand()
	keyId, err := NewUid()
	if err != nil {
		return nil, err
	}
	keyInfo := NewKeyInfo(keyId, key[:])
	return &keyInfo, nil
}

func GenerateUserKey() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	encodeKey, err := encodedPrivateKey(key)
	if err != nil {
		return "", err
	}
	return encodeKey, nil
}
