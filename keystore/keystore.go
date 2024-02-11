package keystore

import (
	"crypto/rsa"
	"ctb-cli/types"
	"fmt"
	"golang.org/x/crypto/curve25519"
)

type Key = types.Key

// KeyStore represents a key store
type KeyStore struct {
	clintId       string
	rootKey       Key
	privateKey    []byte
	recoveryItems []StoreRecoveryItem
	persist       Persist
}

type StoreRecoveryItem struct {
	publicKey *rsa.PublicKey
	sha1      string
}

// Persist KeyStorePersist is an interface for persisting keys
type Persist interface {
	SaveDataKey(keyId string, key string) error
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
		recoveryItems: make([]StoreRecoveryItem, 0),
	}
}

// Insert inserts a key into the key store
func (ks *KeyStore) Insert(keyID string, key Key) error {
	if err := ks.LoadKeys(); err != nil {
		return fmt.Errorf("cannot load keys: %v", err)
	}
	return ks.persistKey(keyID, key)
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
	key, err := ks.OpenDataKey(sk, ks.privateKey)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// persistKey handles the logic of persisting a key
func (ks *KeyStore) persistKey(keyID string, key Key) error {
	// Implement serialization and hashing logic
	pk, err := ks.getPublicKey()
	if err != nil {
		return err
	}
	keyHashed, err := ks.SealDataKey(key[:], pk)
	if err != nil {
		return err
	}

	return ks.persist.SaveDataKey(keyID, keyHashed)
}

func (ks *KeyStore) getPublicKey() ([]byte, error) {
	return curve25519.X25519(ks.privateKey, curve25519.Basepoint)
}
