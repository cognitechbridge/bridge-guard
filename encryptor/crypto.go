package encryptor

import (
	"ctb-cli/types"
	"errors"
	"golang.org/x/crypto/chacha20poly1305"
)

type Crypto struct {
	key   Key
	nonce Nonce
}

func NewCrypto(key Key, nonce Nonce) Crypto {
	return Crypto{key: key, nonce: nonce}
}

type Key = types.Key

// Nonce represents a nonce for ChaCha20-Poly1305.
type Nonce [chacha20poly1305.NonceSize]byte

func GetOverHeadSize() int {
	return chacha20poly1305.Overhead
}

func GetAlgorithmName() string {
	return "AEAD_ChaCha20_Poly1305"
}

// seal encrypts a chunk of data using the ChaCha20-Poly1305 algorithm.
func (c *Crypto) seal(plaintext []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(c.key[:])
	if err != nil {
		return nil, errors.New("failed to create cipher: " + err.Error())
	}

	ciphertext := aead.Seal(nil, c.nonce[:], plaintext, nil)
	return ciphertext, nil
}

// open decrypts a chunk of data using the ChaCha20-Poly1305 algorithm.
func (c *Crypto) open(ciphertext []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(c.key[:])
	if err != nil {
		return nil, errors.New("failed to create cipher: " + err.Error())
	}

	plaintext, err := aead.Open(nil, c.nonce[:], ciphertext, nil)
	if err != nil {
		return nil, errors.New("decryption error: " + err.Error())
	}
	return plaintext, nil
}
