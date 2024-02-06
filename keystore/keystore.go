package keystore

import (
	"crypto/rsa"
	"ctb-cli/types"
)

type Key = types.Key

// KeyStore represents a key store
type KeyStore struct {
	rootKey       Key
	recoveryItems []StoreRecoveryItem
	persist       Persist
}

type StoreRecoveryItem struct {
	publicKey *rsa.PublicKey
	sha1      string
}

// Persist KeyStorePersist is an interface for persisting keys
type Persist interface {
	SaveKey(serializedKey types.SerializedKey) error
	GetKey(keyID string) (*types.SerializedKey, error)
}

// NewKeyStore creates a new instance of KeyStore
func NewKeyStore(rootKey Key, persist Persist) *KeyStore {
	return &KeyStore{
		rootKey:       rootKey,
		persist:       persist,
		recoveryItems: make([]StoreRecoveryItem, 0),
	}
}

// Insert inserts a key into the key store
func (ks *KeyStore) Insert(keyID string, key Key) error {
	return ks.persistKey(keyID, key)
}

// Get retrieves a key from the key store
func (ks *KeyStore) Get(keyID string) (*Key, error) {
	sk, err := ks.persist.GetKey(keyID)
	if err != nil {
		return nil, err
	}

	if sk == nil {
		return nil, err
	}

	key, err := ks.DeserializeKeyPair(sk.Nonce, sk.Key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// persistKey handles the logic of persisting a key
func (ks *KeyStore) persistKey(keyID string, key Key) error {
	// Implement serialization and hashing logic
	nonceHashed, keyHashed, err := ks.SerializeKeyPair(key[:])
	if err != nil {
		return err
	}

	sk := types.SerializedKey{
		ID:    keyID,
		Nonce: nonceHashed,
		Key:   keyHashed,
	}

	return ks.persist.SaveKey(sk)
}
