package core

import (
	"crypto/rand"
	"errors"
	"github.com/btcsuite/btcutil/base58"
	"io"
)

var (
	InvalidPublicKey = errors.New("invalid public key")
)

func EncodePublic(byte []byte) (string, error) {
	return base58.Encode(byte), nil
}

func DecodePublic(str string) ([]byte, error) {
	pub := base58.Decode(str)
	return pub, nil
}

func NewUid() (string, error) {
	rnd := make([]byte, 32)
	_, _ = io.ReadFull(rand.Reader, rnd)
	return EncodeUid(rnd)
}

func EncodeUid(uid []byte) (string, error) {
	return base58.Encode(uid), nil
}
