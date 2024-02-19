package bech32

import "errors"

var (
	InvalidPublicKey = errors.New("invalid public key")
	UidHrp           = "CTB-UID-"
	PubHrp           = "CTB-PUB"
)

func EncodePublic(byte []byte) (string, error) {
	return Encode(PubHrp, byte)
}

func DecodePublic(str string) ([]byte, error) {
	hrp, pub, err := Decode(str)
	if err != nil || hrp != PubHrp {
		return nil, InvalidPublicKey
	}
	return pub, nil
}

func EncodeUid(uid []byte) (string, error) {
	return Encode(UidHrp, uid)
}
