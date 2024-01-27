package keystore

import (
	"crypto/rand"
	"ctb-cli/encryptor"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/chacha20poly1305"
)

// SerializeKeyPair encrypts and serializes the key pair
func (ks *KeyStore) SerializeKeyPair(key []byte) (string, string, error) {
	aead, err := chacha20poly1305.NewX(ks.rootKey[:])
	if err != nil {
		return "", "", fmt.Errorf("failed to create cipher: %w", err)
	}

	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	_, err = rand.Read(nonce)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphered := aead.Seal(nil, nonce, key, nil)
	nonceEncoded := base64.StdEncoding.EncodeToString(nonce)
	cipheredEncoded := base64.StdEncoding.EncodeToString(ciphered)

	return nonceEncoded, cipheredEncoded, nil
}

// DeserializeKeyPair decrypts and deserializes the key pair
func (ks *KeyStore) DeserializeKeyPair(nonceEncoded, cipheredEncoded string) (*encryptor.Key, error) {
	nonce, err := base64.StdEncoding.DecodeString(nonceEncoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode nonce: %w", err)
	}

	ciphered, err := base64.StdEncoding.DecodeString(cipheredEncoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphered data: %w", err)
	}

	aead, err := chacha20poly1305.NewX(ks.rootKey[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	deciphered, err := aead.Open(nil, nonce, ciphered, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	key := encryptor.Key{}
	copy(key[:], deciphered)

	return &key, nil
}
