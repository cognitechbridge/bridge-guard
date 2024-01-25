package encryptor

import (
	"errors"
	"golang.org/x/crypto/chacha20poly1305"
)

// Key represents a 256-bit key used for ChaCha20-Poly1305.
type Key [chacha20poly1305.KeySize]byte

// Nonce represents a nonce for ChaCha20-Poly1305.
type Nonce [chacha20poly1305.NonceSize]byte

func GetOverHeadSize() int {
	return chacha20poly1305.Overhead
}

func GetAlgorithmName() string {
	return "ChaChaPoly1350"
}

// EncryptChunk encrypts a chunk of data using the ChaCha20-Poly1305 algorithm.
func encryptChunk(plaintext []byte, key Key, nonce Nonce) ([]byte, error) {
	aead, err := chacha20poly1305.New(key[:])
	if err != nil {
		return nil, errors.New("failed to create cipher: " + err.Error())
	}

	ciphertext := aead.Seal(nil, nonce[:], plaintext, nil)
	return ciphertext, nil
}

// DecryptChunk decrypts a chunk of data using the ChaCha20-Poly1305 algorithm.
func decryptChunk(ciphertext []byte, key Key, nonce Nonce) ([]byte, error) {
	aead, err := chacha20poly1305.New(key[:])
	if err != nil {
		return nil, errors.New("failed to create cipher: " + err.Error())
	}

	plaintext, err := aead.Open(nil, nonce[:], ciphertext, nil)
	if err != nil {
		return nil, errors.New("decryption error: " + err.Error())
	}
	return plaintext, nil
}
