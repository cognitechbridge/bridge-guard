package core

import (
	"encoding/json"
	"errors"
)

var (
	ErrKeyNotFoundInVault = errors.New("key not found in vault")
)

type Vault struct {
	Id       string `json:"id"`
	KeyId    string `json:"keyId"`
	ParentId string `json:"parentId"`
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
