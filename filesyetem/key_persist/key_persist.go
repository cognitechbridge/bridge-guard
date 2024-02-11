package key_persist

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type KeyPersist struct {
	clientId string
	rootPath string
}

func New(clientId string, rootPath string) *KeyPersist {
	return &KeyPersist{
		rootPath: rootPath,
		clientId: clientId,
	}
}

func (k *KeyPersist) getDataPath() string {
	p := filepath.Join(k.rootPath, "keys", "data", k.clientId)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		os.MkdirAll(p, os.ModePerm)
	}
	return p
}

func (k *KeyPersist) getPrivatePath() string {
	p := filepath.Join(k.rootPath, "keys", "private")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		os.MkdirAll(p, os.ModePerm)
	}
	return p
}

func (k *KeyPersist) getPublicPath() string {
	p := filepath.Join(k.rootPath, "keys", "public")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		os.MkdirAll(p, os.ModePerm)
	}
	return p
}

func (k *KeyPersist) GetPublicKey(id string) (*rsa.PublicKey, error) {
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

func (k *KeyPersist) SavePublicKey(id string, key string) (err error) {
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

func (k *KeyPersist) GetPrivateKey() (string, error) {
	p := filepath.Join(k.getPrivatePath(), k.clientId)
	content, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (k *KeyPersist) SavePrivateKey(key string) (err error) {
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

func (k *KeyPersist) SaveDataKey(keyId string, key string) error {
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

func (k *KeyPersist) GetDataKey(keyID string) (string, error) {
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
