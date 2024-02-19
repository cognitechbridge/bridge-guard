package types

import (
	"crypto/rand"
	"ctb-cli/crypto/bech32"
	"errors"
	"io"
)

type Recipient struct {
	Email  string `json:"email,omitempty"`
	Public string `json:"public,omitempty"`
	UserId string `json:"userId,omitempty"`
}

var (
	InvalidRecipient         = errors.New("invalid recipient address")
	ErrorGeneratingRecipient = errors.New("error generating recipient")

	bech32pre = "ctb-pub"
)

func NewRecipient(email string, public []byte, userId string) (Recipient, error) {
	publicStr, err := bech32.Encode(bech32pre, public)
	if err != nil {
		return Recipient{}, err
	}
	return Recipient{Email: email, Public: publicStr, UserId: userId}, nil
}

func GenerateRandomRecipient() (string, error) {
	add := make([]byte, 32)
	io.ReadFull(rand.Reader, add)
	res, err := bech32.Encode(bech32pre, add)
	if err != nil {
		return "", ErrorGeneratingRecipient
	}
	return res, nil
}

func (r Recipient) GetPublicBytes() ([]byte, error) {
	hrp, data, err := bech32.Decode(r.Public)
	if err != nil || hrp != bech32pre {
		return nil, InvalidRecipient
	}
	return data, nil
}
