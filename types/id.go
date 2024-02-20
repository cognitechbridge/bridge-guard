package types

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"io"
	"strings"
)

var (
	InvalidPublicKey = errors.New("invalid public key")
)

// Crockford's Base32 Alphabet in lower case
const crockfordAlphabetLowerCase = "0123456789abcdefghjkmnpqrstvwxyz"

// Custom encoding using Crockford's lower case alphabet.
var crockfordEncoding = base32.NewEncoding(crockfordAlphabetLowerCase).WithPadding(base32.NoPadding)

// encodeCrockford encodes the given data to Crockford's Base32 using lower case.
func encodeCrockford(data []byte) string {
	return crockfordEncoding.EncodeToString(data)
}

// decodeCrockford decodes the given Crockford's Base32 encoded string in lower case.
// It handles the common character confusions and omits, and is case insensitive.
func decodeCrockford(s string) ([]byte, error) {
	// Prepare string for decoding: map easily confused characters to the expected ones,
	// and convert to lower case to match our custom lower case alphabet.
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "o", "0")
	s = strings.ReplaceAll(s, "i", "1")
	s = strings.ReplaceAll(s, "l", "1")

	return crockfordEncoding.DecodeString(s)
}

func EncodePublic(byte []byte) (string, error) {
	return encodeCrockford(byte), nil
}

func DecodePublic(str string) ([]byte, error) {
	pub, err := decodeCrockford(str)
	if err != nil {
		return nil, InvalidPublicKey
	}
	return pub, nil
}

func NewUid() (string, error) {
	rnd := make([]byte, 32)
	_, _ = io.ReadFull(rand.Reader, rnd)
	return EncodeUid(rnd)
}

func EncodeUid(uid []byte) (string, error) {
	return encodeCrockford(uid), nil
}
