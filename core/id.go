package core

import (
	"crypto/rand"
	"errors"
	"io"

	"github.com/btcsuite/btcutil/base58"
)

var (
	InvalidPublicKey = errors.New("invalid public key")
)

func encodedPrivateKey(privateKey []byte) (string, error) {
	return base58.Encode(privateKey), nil
}

func EncodePublic(byte []byte) (string, error) {
	return base58.Encode(byte), nil
}

func DecodePublic(str string) ([]byte, error) {
	pub := base58.Decode(str)
	return pub, nil
}

func DecodePrivateKey(str string) ([]byte, error) {
	return base58.Decode(str), nil
}

func NewUid() (string, error) {
	rnd := make([]byte, 32)
	_, _ = io.ReadFull(rand.Reader, rnd)
	return EncodeUid(rnd)
}

func EncodeUid(uid []byte) (string, error) {
	return base58.Encode(uid), nil
}
