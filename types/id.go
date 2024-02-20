package types

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

var (
	InvalidPublicKey = errors.New("invalid public key")
)

func EncodePublic(byte []byte) (string, error) {
	return hex.EncodeToString(byte), nil
}

func DecodePublic(str string) ([]byte, error) {
	pub, err := hex.DecodeString(str)
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
	return hex.EncodeToString(uid), nil
}
