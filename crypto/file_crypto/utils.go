package file_crypto

import (
	"ctb-cli/types"
	"golang.org/x/crypto/chacha20poly1305"
	"math/big"
)

const (
	lastChunkFlag = 0x01
)

// nonce represents a key for ChaCha20-Poly1305.
type key = types.Key

// nonce represents a nonce for ChaCha20-Poly1305.
type nonce [chacha20poly1305.NonceSize]byte

func getOverHeadSize() int {
	return chacha20poly1305.Overhead
}

func getAlgorithmName() string {
	return "AEAD_ChaCha20_Poly1305"
}

func (nc *nonce) setLastChunkFlag() {
	nc[len(nc)-1] = lastChunkFlag
}

func (nc *nonce) increaseBe() {
	number := new(big.Int).SetBytes(nc[:])
	number.Add(number, big.NewInt(1))
	newBytes := number.Bytes()
	copy(nc[len(nc)-len(newBytes):], newBytes)
}
