package recovery

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"ctb-cli/types"
	"github.com/google/uuid"

	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
)

type Recovery struct {
	Version string `json:"version"`
	Alg     string `json:"alg"`
	Cipher  string `json:"cipher"`
	Sha1    string `json:"sha1"`
}

func generateRecoveryBlob(key types.Key, recoveryItems []types.RecoveryItem) ([]string, error) {
	recoveryList := recoveryItems
	if recoveryList == nil || len(recoveryList) == 0 {
		return nil, fmt.Errorf("recoveryItems key not found. Cannot generate data key")
	}
	blobs := make([]string, 0)
	for _, rec := range recoveryList {
		encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rec.PublicKey, key[:], nil)
		if err != nil {
			return nil, err
		}
		recovery := Recovery{
			Version: "V1",
			Alg:     "RSAES_OAEP_SHA_256",
			Cipher:  base64.StdEncoding.EncodeToString(encrypted),
			Sha1:    rec.Sha1,
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

func GenerateKey(recoveryItems []types.RecoveryItem) (*types.KeyInfo, error) {
	key := types.Key{}
	if _, err := io.ReadFull(rand.Reader, key[:]); err != nil {
		return nil, err
	}

	keyId, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	blobs, err := generateRecoveryBlob(key, recoveryItems)
	if err != nil {
		return nil, err
	}

	return &types.KeyInfo{
		Key:           key[:],
		Id:            keyId.String(),
		RecoveryBlobs: blobs,
	}, nil
}
