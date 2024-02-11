package filesyetem

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
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

func (k *KeyStoreFilesystem) getDataPath() string {
	root, _ := GetRepoCtbRoot()
	p := filepath.Join(root, "keys", "data", k.clientId)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		os.MkdirAll(p, os.ModePerm)
	}
	return p
}

func (k *KeyStoreFilesystem) getPrivatePath() string {
	root, _ := GetRepoCtbRoot()
	p := filepath.Join(root, "keys", "private")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		os.MkdirAll(p, os.ModePerm)
	}
	return p
}

func (k *KeyStoreFilesystem) getPublicPath() string {
	root, _ := GetRepoCtbRoot()
	p := filepath.Join(root, "keys", "public")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		os.MkdirAll(p, os.ModePerm)
	}
	return p
}

func (k *KeyStoreFilesystem) GetPublicKey(id string) (*rsa.PublicKey, error) {
	p := filepath.Join(k.getPublicPath(), id)
	file, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(file)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the key")
	}
	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return publicKey, nil
}

func (k *KeyStoreFilesystem) SavePublicKey(id string, key string) (err error) {
	p := filepath.Join(k.getPublicPath(), id)

	file, err := os.Create(p)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(key)
	if err != nil {
		return err
	}
	return nil
}

func (k *KeyStoreFilesystem) GetPrivateKey() (string, error) {
	p := filepath.Join(k.getPrivatePath(), k.clientId)
	content, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (k *KeyStoreFilesystem) SavePrivateKey(key string) (err error) {
	p := filepath.Join(k.getPrivatePath(), k.clientId)
	file, err := os.Create(p)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(key)
	if err != nil {
		return err
	}
	return nil
}

func (k *KeyStoreFilesystem) SaveDataKey(keyId string, key string) error {
	p := filepath.Join(k.getDataPath(), keyId)
	file, err := os.Create(p)
	defer file.Close()
	if err != nil {
		return err
	}
	_, err = file.Write([]byte(key))
	if err != nil {
		return err
	}
	return nil
}

func (k *KeyStoreFilesystem) GetDataKey(keyID string) (string, error) {
	p := filepath.Join(k.getDataPath(), keyID)
	file, err := os.Open(p)
	defer file.Close()
	if err != nil {
		return "", err
	}
	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(content), err
}
