package core

import (
	"encoding/json"
	"errors"
)

var (
	ErrorKeyNotFoundInVault = errors.New("key not found in vault")
)

type Vault struct {
	Id            string                      `json:"id"`
	KeyId         string                      `json:"keyId"`
	EncryptedKeys map[string]EncryptedDataKey `json:"encryptedKeys"`
}

type EncryptedDataKey struct {
	KeyId          string
	EncryptedValue string
}

func (v *Vault) Serialize() ([]byte, error) {
	return json.Marshal(v)
}

func UnmarshalVault(data []byte) (Vault, error) {
	var vault Vault
	err := json.Unmarshal(data, &vault)
	if err != nil {
		return Vault{}, err
	}
	return vault, nil
}

func (v *Vault) AddKey(sealed string, keyId string) error {
	e := EncryptedDataKey{
		KeyId:          keyId,
		EncryptedValue: sealed,
	}
	v.EncryptedKeys[keyId] = e
	return nil
}
