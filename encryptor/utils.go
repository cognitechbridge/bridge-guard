package encryptor

import (
	"bytes"
	"ctb-cli/types"
	"golang.org/x/crypto/chacha20poly1305"
	"math/big"
)

const (
	lastChunkFlag = 0x01
)

type Key = types.Key

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

// isZeroes checks if all bytes in the buffer are zero.
func isZeroes(buf []byte) bool {
	// Create a slice of zeroes with the same length as buf
	zeroes := make([]byte, len(buf))
	return bytes.Equal(buf, zeroes)
}
