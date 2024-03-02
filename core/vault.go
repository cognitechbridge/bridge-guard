package core

import (
	"encoding/json"
	"errors"
)

var (
	ErrorKeyNotFoundInVault = errors.New("key not found in vault")
)

type Vault struct {
	Id            string            `json:"id"`
	KeyId         string            `json:"keyId"`
	EncryptedKeys map[string]string `json:"encryptedKeys"`
	ParentId      string            `json:"parentId"`
}

func (v *Vault) Marshal() ([]byte, error) {
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
	v.EncryptedKeys[keyId] = sealed
	return nil
}
