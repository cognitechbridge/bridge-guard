package keystore

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"ctb-cli/encryptor"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
)

type RecoveryVersion int

type GeneratedKey struct {
	Key          encryptor.Key
	RecoveryBlob string
}

type Recovery struct {
	Version      string `json:"version"`
	Alg          string `json:"alg"`
	Cipher       string `json:"cipher"`
	RecoverySha1 string `json:"recoverySha1"`
}

func (ks *KeyStore) GenerateRecoveryBlob(key encryptor.Key) (string, error) {
	recoveryKey, err := ks.GetRecoveryKey()
	if err != nil || recoveryKey == nil {
		return "", fmt.Errorf("recovery key not found. Cannot generate data key")
	}

	encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, recoveryKey, key[:], nil)
	if err != nil {
		return "", err
	}

	recovery := Recovery{
		Version:      "V1",
		Alg:          "RSAES_OAEP_SHA_256",
		Cipher:       base64.StdEncoding.EncodeToString(encrypted),
		RecoverySha1: ks.recoverySha1,
	}

	serialized, err := json.Marshal(recovery)
	if err != nil {
		return "", err
	}

	blob := base64.StdEncoding.EncodeToString(serialized)
	return blob, nil
}

func (ks *KeyStore) GenerateKeyPair(keyId string) (GeneratedKey, error) {
	key := encryptor.Key{}
	if _, err := io.ReadFull(rand.Reader, key[:]); err != nil {
		return GeneratedKey{}, err
	}

	blob, err := ks.GenerateRecoveryBlob(key)
	if err != nil {
		return GeneratedKey{}, err
	}

	err = ks.Insert(keyId, key)
	if err != nil {
		return GeneratedKey{}, err
	}

	return GeneratedKey{
		Key:          key,
		RecoveryBlob: blob,
	}, nil
}

func (ks *KeyStore) GetRecoveryKey() (*rsa.PublicKey, error) {
	return ks.recoveryPublicKey, nil
}
