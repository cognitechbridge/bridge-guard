package keystore

import (
	"crypto/rsa"
	"ctb-cli/crypto/key_crypto"
	"ctb-cli/crypto/recovery"
	"ctb-cli/types"
	"fmt"
	"golang.org/x/crypto/curve25519"
	"os"
)

type Key = types.Key

// KeyStore represents a key store
type KeyStore struct {
	clintId       string
	rootKey       Key
	privateKey    []byte
	recoveryItems []types.RecoveryItem
	persist       Persist
}

// Persist KeyStorePersist is an interface for persisting keys
type Persist interface {
	SaveDataKey(keyId, key, recipient string) error
	GetDataKey(keyID string) (string, error)
	GetPrivateKey() (string, error)
	SavePrivateKey(key string) (err error)
	GetPublicKey(id string) (*rsa.PublicKey, error)
	SavePublicKey(id string, key string) (err error)
}

// NewKeyStore creates a new instance of KeyStore
func NewKeyStore(clientId string, rootKey Key, persist Persist) *KeyStore {
	return &KeyStore{
		clintId:       clientId,
		rootKey:       rootKey,
		persist:       persist,
		recoveryItems: make([]types.RecoveryItem, 0),
	}
}

// Insert inserts a key into the key store
func (ks *KeyStore) Insert(keyID string, key Key) error {
	if err := ks.LoadKeys(); err != nil {
		return fmt.Errorf("cannot load keys: %v", err)
	}

	pk, err := ks.getPublicKey()
	if err != nil {
		return err
	}
	keyHashed, err := key_crypto.SealDataKey(key[:], pk)
	if err != nil {
		return err
	}

	return ks.persist.SaveDataKey(keyID, keyHashed, ks.clintId)
}

// Get retrieves a key from the key store
func (ks *KeyStore) Get(keyID string) (*Key, error) {
	if err := ks.LoadKeys(); err != nil {
		return nil, fmt.Errorf("cannot load keys: %v", err)
	}
	sk, err := ks.persist.GetDataKey(keyID)
	if err != nil {
		return nil, err
	}
	key, err := key_crypto.OpenDataKey(sk, ks.privateKey)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (ks *KeyStore) Share(keyId string, recipient []byte, recipientClientId string) error {
	if err := ks.LoadKeys(); err != nil {
		return fmt.Errorf("cannot load keys: %v", err)
	}

	key, err := ks.Get(keyId)
	if err != nil {
		return fmt.Errorf("cannot load key: %v", err)
	}

	keyHashed, err := key_crypto.SealDataKey(key[:], recipient)
	if err != nil {
		return err
	}

	return ks.persist.SaveDataKey(keyId, keyHashed, recipientClientId)
}

func (ks *KeyStore) GetRecoveryItems() ([]types.RecoveryItem, error) {
	return ks.recoveryItems, nil
}

func (ks *KeyStore) AddRecoveryKey(inPath string) error {
	path := os.ExpandEnv(inPath)
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	rec, err := recovery.UnmarshalRecoveryItem(pemBytes)
	if err != nil {
		return err
	}

	ks.recoveryItems = append(ks.recoveryItems, *rec)

	return nil
}

func (ks *KeyStore) getPublicKey() ([]byte, error) {
	return curve25519.X25519(ks.privateKey, curve25519.Basepoint)
}
