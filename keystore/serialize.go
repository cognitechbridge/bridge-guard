package keystore

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"ctb-cli/types"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
	"io"
	"strings"
)

const (
	info = "cognitechbridge.com/v1/X25519"
)

// DeserializePrivateKey encrypts and serializes the private key
func (*KeyStore) DeserializePrivateKey(serialized []byte, rootKey *types.Key) ([]byte, error) {
	parts := strings.Split(string(serialized), "\n")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid serialized key)")
	}
	salt, err1 := base64.RawStdEncoding.DecodeString(parts[0])
	ciphered, err2 := base64.RawStdEncoding.DecodeString(parts[1])
	if errors.Join(err1, err2) != nil {
		return nil, fmt.Errorf("invalid serialized key")
	}

	derivedKey, err := deriveKey(rootKey, salt)

	aead, err := chacha20poly1305.New(derivedKey[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	nonce := make([]byte, chacha20poly1305.NonceSize)

	deciphered, err := aead.Open(nil, nonce, ciphered, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return deciphered, nil
}

// SerializePrivateKey encrypts and serializes the private key
func (*KeyStore) SerializePrivateKey(privateKey []byte, rootKey *types.Key) ([]byte, error) {
	salt := make([]byte, 16)
	rand.Read(salt)

	derivedKey, err := deriveKey(rootKey, salt)

	aead, err := chacha20poly1305.New(derivedKey[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	nonce := make([]byte, chacha20poly1305.NonceSize)
	ciphered := aead.Seal(nil, nonce, privateKey, nil)

	res := fmt.Sprintf("%s\n%s",
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(ciphered),
	)

	return []byte(res), nil
}

func deriveKey(rootKey *types.Key, salt []byte) (derivedKey types.Key, err error) {
	hkdf := hkdf.New(sha256.New, rootKey[:], salt, []byte(info))
	derivedKey = types.Key{}
	_, err = io.ReadFull(hkdf, derivedKey[:])
	return
}

// SerializeDataKey encrypts and serializes the key pair
func (*KeyStore) SerializeDataKey(key []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, key[:], nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	return encrypted, nil
}

// DeserializeDataKey decrypts and deserializes the key pair
func (*KeyStore) DeserializeDataKey(ciphered []byte, privateKey *rsa.PrivateKey) (*Key, error) {
	deciphered, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphered, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	key := Key{}
	copy(key[:], deciphered)

	return &key, nil
}
