package keystore

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"

	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
)

type RecoveryVersion int

type GeneratedKey struct {
	Key           Key
	RecoveryBlobs []string
}

type Recovery struct {
	Version string `json:"version"`
	Alg     string `json:"alg"`
	Cipher  string `json:"cipher"`
	Sha1    string `json:"sha1"`
}

func (ks *KeyStore) GenerateRecoveryBlob(key Key) ([]string, error) {
	recoveryList := ks.recoveryItems
	if recoveryList == nil || len(recoveryList) == 0 {
		return nil, fmt.Errorf("recoveryItems key not found. Cannot generate data key")
	}
	blobs := make([]string, 0)
	for _, rec := range recoveryList {
		encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rec.publicKey, key[:], nil)
		if err != nil {
			return nil, err
		}
		recovery := Recovery{
			Version: "V1",
			Alg:     "RSAES_OAEP_SHA_256",
			Cipher:  base64.StdEncoding.EncodeToString(encrypted),
			Sha1:    rec.sha1,
		}
		serialized, err := json.Marshal(recovery)
		if err != nil {
			return nil, err
		}
		blob := base64.StdEncoding.EncodeToString(serialized)
		blobs = append(blobs, blob)
	}
	return blobs, nil
}

func (ks *KeyStore) GenerateKeyPair(keyId string) (GeneratedKey, error) {
	key := Key{}
	if _, err := io.ReadFull(rand.Reader, key[:]); err != nil {
		return GeneratedKey{}, err
	}

	blobs, err := ks.GenerateRecoveryBlob(key)
	if err != nil {
		return GeneratedKey{}, err
	}

	err = ks.Insert(keyId, key)
	if err != nil {
		return GeneratedKey{}, err
	}

	return GeneratedKey{
		Key:           key,
		RecoveryBlobs: blobs,
	}, nil
}
