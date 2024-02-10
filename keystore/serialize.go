package keystore

import (
	"crypto/rand"
	"crypto/sha256"
	"ctb-cli/types"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
	"io"
	"strings"
)

const (
	info = "cognitechbridge.com/v1/X25519"
)

// DeserializePrivateKey encrypts and serializes the private key
func (*KeyStore) DeserializePrivateKey(serialized string, rootKey *types.Key) ([]byte, error) {
	parts := strings.Split(serialized, "\n")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid serialized key)")
	}
	salt, err1 := base64.RawStdEncoding.DecodeString(parts[0])
	ciphered, err2 := base64.RawStdEncoding.DecodeString(parts[1])
	if errors.Join(err1, err2) != nil {
		return nil, fmt.Errorf("invalid serialized key")
	}

	derivedKey, err := deriveKey(rootKey[:], salt)

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
func (*KeyStore) SerializePrivateKey(privateKey []byte, rootKey *types.Key) (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("error generating random salt: %v", err)
	}

	derivedKey, err := deriveKey(rootKey[:], salt)

	aead, err := chacha20poly1305.New(derivedKey[:])
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	nonce := make([]byte, chacha20poly1305.NonceSize)
	ciphered := aead.Seal(nil, nonce, privateKey, nil)

	res := fmt.Sprintf("%s\n%s",
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(ciphered),
	)

	return res, nil
}

func deriveKey(rootKey []byte, salt []byte) (derivedKey types.Key, err error) {
	hk := hkdf.New(sha256.New, rootKey[:], salt, []byte(info))
	derivedKey = types.Key{}
	_, err = io.ReadFull(hk, derivedKey[:])
	return
}

// SerializeDataKey encrypts and serializes the key pair
func (*KeyStore) SerializeDataKey(key []byte, publicKey []byte) (string, error) {
	ephemeralSecret := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, ephemeralSecret[:])
	if err != nil {
		return "", fmt.Errorf("error generating random ephemeral secret: %v", err)
	}

	ephemeralShare, err := curve25519.X25519(ephemeralSecret, curve25519.Basepoint)
	if err != nil {
		return "", fmt.Errorf("error encrypting data key: %v", err)
	}

	ephemeralShareString := base64.RawStdEncoding.EncodeToString(ephemeralShare)
	publicKeyString := base64.RawStdEncoding.EncodeToString(publicKey)
	salt := ephemeralShareString + publicKeyString

	sharedSecret, err := curve25519.X25519(ephemeralSecret, publicKey)
	if err != nil {
		return "", fmt.Errorf("error encrypting data key: %v", err)
	}

	wrapKey, err := deriveKey(sharedSecret, []byte(salt))

	aead, err := chacha20poly1305.New(wrapKey[:])
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w\n", err)
	}

	nonce := make([]byte, chacha20poly1305.NonceSize)
	ciphered := aead.Seal(nil, nonce, key, nil)

	res := fmt.Sprintf("%s\n%s",
		ephemeralShareString,
		base64.RawStdEncoding.EncodeToString(ciphered),
	)
	return res, nil
}

// DeserializeDataKey decrypts and deserializes the key pair
func (*KeyStore) DeserializeDataKey(serialized string, privateKey []byte) (*Key, error) {
	parts := strings.Split(serialized, "\n")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid serialized key)")
	}
	ephemeralShareString := parts[0]
	ephemeralShare, err1 := base64.RawStdEncoding.DecodeString(ephemeralShareString)
	ciphered, err2 := base64.RawStdEncoding.DecodeString(parts[1])
	if errors.Join(err1, err2) != nil {
		return nil, fmt.Errorf("invalid serialized key")
	}

	publicKey, err := derivePublicFromPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting data key: %v", err)
	}
	salt := ephemeralShareString + base64.RawStdEncoding.EncodeToString(publicKey)

	sharedSecret, err := curve25519.X25519(privateKey, ephemeralShare)
	if err != nil {
		return nil, fmt.Errorf("error decrypting data key: %v", err)
	}

	wrapKey, err := deriveKey(sharedSecret, []byte(salt))

	aead, err := chacha20poly1305.New(wrapKey[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	nonce := make([]byte, chacha20poly1305.NonceSize)

	deciphered, err := aead.Open(nil, nonce, ciphered, nil)
	if err != nil {
		return nil, fmt.Errorf("error decrypting data key: %v", err)
	}

	key := Key{}
	copy(key[:], deciphered)

	return &key, nil
}

func derivePublicFromPrivateKey(privateKey []byte) ([]byte, error) {
	return curve25519.X25519(privateKey, curve25519.Basepoint)
}
