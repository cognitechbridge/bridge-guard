package core

import "encoding/json"

type Vault struct {
	Id            string
	KeyId         string
	EncryptedKeys []string
}

type EncryptedKey struct {
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
