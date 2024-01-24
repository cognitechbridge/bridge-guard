package keystore

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/google/uuid"
	"golang.org/x/crypto/chacha20poly1305"
	"io"
	"storage-go/encryptor"
)

const RecoveryTag = "RECOVERY"

type RecoveryVersion int

type GeneratedKey struct {
	Key          encryptor.Key
	RecoveryBlob string
}

type Recovery struct {
	Version string
	Alg     string
	Nonce   string
	Cipher  string
	ID      string
}

func (ks *KeyStore) GenerateRecoveryBlob(key encryptor.Key, nonce []byte) (string, error) {
	recoveryID, recoveryKey, err := ks.GetRecoveryKey()
	if err != nil {
		return "", err
	}

	aead, err := chacha20poly1305.NewX(recoveryKey[:])
	if err != nil {
		return "", err
	}

	encrypted := aead.Seal(nil, nonce[:], key[:], nil)

	recovery := Recovery{
		Version: "V1",
		Alg:     "XChaCha20Poly1305",
		Nonce:   base64.StdEncoding.EncodeToString(nonce[:]),
		Cipher:  base64.StdEncoding.EncodeToString(encrypted),
		ID:      recoveryID,
	}

	serialized, err := json.Marshal(recovery)
	if err != nil {
		return "", err
	}

	blob := base64.StdEncoding.EncodeToString(serialized)
	return blob, nil
}

func (ks *KeyStore) GenerateKeyPair() (GeneratedKey, error) {
	key := encryptor.Key{}
	if _, err := io.ReadFull(rand.Reader, key[:]); err != nil {
		return GeneratedKey{}, err
	}

	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return GeneratedKey{}, err
	}

	blob, err := ks.GenerateRecoveryBlob(key, nonce)
	if err != nil {
		return GeneratedKey{}, err
	}

	keyId := uuid.New()
	err = ks.Insert(keyId.String(), key)
	if err != nil {
		return GeneratedKey{}, err
	}

	return GeneratedKey{
		Key:          key,
		RecoveryBlob: blob,
	}, nil
}

func (ks *KeyStore) GetRecoveryKey() (string, encryptor.Key, error) {
	return ks.GetWithTag(RecoveryTag)
}
