package key_repository

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type KeyRepository struct {
	clientId string
	rootPath string
}

func New(clientId string, rootPath string) *KeyRepository {
	return &KeyRepository{
		rootPath: rootPath,
		clientId: clientId,
	}
}

func (k *KeyRepository) GetPublicKey(id string) (*rsa.PublicKey, error) {
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

func (k *KeyRepository) SavePublicKey(id string, key string) (err error) {
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

func (k *KeyRepository) GetPrivateKey() (string, error) {
	p := filepath.Join(k.getPrivatePath(), k.clientId)
	content, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (k *KeyRepository) SavePrivateKey(key string) (err error) {
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

func (k *KeyRepository) SaveDataKey(keyId, key, recipient string) error {
	p := filepath.Join(k.getDataPath(recipient), keyId)
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

func (k *KeyRepository) GetDataKey(keyID string) (string, error) {
	p := filepath.Join(k.getDataPath(k.clientId), keyID)
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

func (k *KeyRepository) getDataPath(recipient string) string {
	p := filepath.Join(k.rootPath, "keys", "data", recipient)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		os.MkdirAll(p, os.ModePerm)
	}
	return p
}

func (k *KeyRepository) getPrivatePath() string {
	p := filepath.Join(k.rootPath, "keys", "private")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		os.MkdirAll(p, os.ModePerm)
	}
	return p
}

func (k *KeyRepository) getPublicPath() string {
	p := filepath.Join(k.rootPath, "keys", "public")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		os.MkdirAll(p, os.ModePerm)
	}
	return p
}
