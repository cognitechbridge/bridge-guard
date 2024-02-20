package bech32

import (
	"errors"
	"github.com/btcsuite/btcutil/base58"
)

var (
	InvalidPublicKey = errors.New("invalid public key")
)

func EncodePublic(byte []byte) (string, error) {
	return base58.CheckEncode(byte, 1), nil
}

func DecodePublic(str string) ([]byte, error) {
	pub, version, err := base58.CheckDecode(str)
	if err != nil || version != 1 {
		return nil, InvalidPublicKey
	}
	return pub, nil
}

func EncodeUid(uid []byte) (string, error) {
	return base58.Encode(uid), nil
}
