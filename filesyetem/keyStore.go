package filesyetem

import (
	"ctb-cli/types"
	"io"
	"os"
	"path/filepath"
)

type KeyStoreFilesystem struct {
	clientId string
}

func NewKeyStoreFilesystem(clientId string) *KeyStoreFilesystem {
	return &KeyStoreFilesystem{
		clientId: clientId,
	}
}

func (k *KeyStoreFilesystem) getPath() string {
	root, _ := GetRepoCtbRoot()
	p := filepath.Join(root, "keys", k.clientId)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		os.MkdirAll(p, os.ModePerm)
	}
	return p
}

func (k *KeyStoreFilesystem) SaveKey(serializedKey types.SerializedKey) error {
	p := filepath.Join(k.getPath(), serializedKey.ID)
	file, err := os.Create(p)
	defer file.Close()
	if err != nil {
		return err
	}
	_, err = file.WriteString(serializedKey.Key)
	if err != nil {
		return err
	}
	return nil
}

func (k *KeyStoreFilesystem) GetKey(keyID string) (*types.SerializedKey, error) {
	p := filepath.Join(k.getPath(), keyID)
	file, err := os.Open(p)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	sk := types.SerializedKey{
		ID:  keyID,
		Key: string(content),
	}
	return &sk, err
}
