package types

import (
	"crypto/rand"
	"errors"
	"io"
)

type Recipient struct {
	Email  string `json:"email,omitempty"`
	Public string `json:"public,omitempty"`
	UserId string `json:"userId,omitempty"`
}

var (
	ErrorGeneratingRecipient = errors.New("error generating recipient")
)

func NewRecipient(email string, public []byte, userId string) (Recipient, error) {
	publicStr, err := EncodePublic(public)
	if err != nil {
		return Recipient{}, err
	}
	return Recipient{Email: email, Public: publicStr, UserId: userId}, nil
}

func GenerateRandomRecipient() (string, error) {
	add := make([]byte, 32)
	io.ReadFull(rand.Reader, add)
	res, err := EncodePublic(add)
	if err != nil {
		return "", ErrorGeneratingRecipient
	}
	return res, nil
}

func (r Recipient) GetPublicBytes() ([]byte, error) {
	data, err := DecodePublic(r.Public)
	return data, err
}
