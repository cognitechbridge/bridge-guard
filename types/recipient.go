package types

import (
	"crypto/rand"
	"ctb-cli/crypto/bech32"
	"errors"
	"io"
)

type Recipient struct {
	Email    string `json:"email,omitempty"`
	Public   string `json:"public,omitempty"`
	ClientId string `json:"clientId,omitempty"`
}

var (
	InvalidRecipient         = errors.New("invalid recipient address")
	ErrorGeneratingRecipient = errors.New("error generating recipient")
)

func NewRecipient(email string, public []byte, clientId string) (Recipient, error) {
	publicStr, err := bech32.Encode("ctb", public)
	if err != nil {
		return Recipient{}, err
	}
	return Recipient{Email: email, Public: publicStr, ClientId: clientId}, nil
}

func GenerateRandomRecipient() (string, error) {
	add := make([]byte, 32)
	io.ReadFull(rand.Reader, add)
	res, err := bech32.Encode("ctb", add)
	if err != nil {
		return "", ErrorGeneratingRecipient
	}
	return res, nil
}

func (r Recipient) GetPublicBytes() ([]byte, error) {
	hrp, data, err := bech32.Decode(r.Public)
	if err != nil || hrp != "ctb" {
		return nil, InvalidRecipient
	}
	return data, nil
}
