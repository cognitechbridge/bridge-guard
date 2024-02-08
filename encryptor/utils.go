package encryptor

import (
	"bytes"
	"math/big"
)

const (
	lastChunkFlag = 0x01
)

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
