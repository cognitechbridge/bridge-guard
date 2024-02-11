package file_crypto

import (
	"golang.org/x/crypto/chacha20poly1305"
	"math/big"
)

const (
	lastChunkFlag = 0x01
)

// Nonce represents a nonce for ChaCha20-Poly1305.
type Nonce [chacha20poly1305.NonceSize]byte

func GetOverHeadSize() int {
	return chacha20poly1305.Overhead
}

func GetAlgorithmName() string {
	return "AEAD_ChaCha20_Poly1305"
}

func (nc *Nonce) setLastChunkFlag() {
	nc[len(nc)-1] = lastChunkFlag
}

func (nc *Nonce) increaseBe() {
	number := new(big.Int).SetBytes(nc[:])
	number.Add(number, big.NewInt(1))
	newBytes := number.Bytes()
	copy(nc[len(nc)-len(newBytes):], newBytes)
}
