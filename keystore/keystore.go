package keystore

import (
	"crypto/rsa"
	"ctb-cli/encryptor"
)

type SerializedKey struct {
	ID    string
	Nonce string
	Key   string
	Tag   string
}

// Persist KeyStorePersist is an interface for persisting keys
type Persist interface {
	SaveKey(serializedKey SerializedKey) error
	GetKey(keyID string) (*SerializedKey, error)
	GetWithTag(tag string) (*SerializedKey, error)
}

// KeyStore represents a key store
type KeyStore struct {
	rootKey           encryptor.Key
	recoveryPublicKey *rsa.PublicKey
	recoverySha1      string
	persist           Persist
}

// NewKeyStore creates a new instance of KeyStore
func NewKeyStore(rootKey encryptor.Key, persist Persist) *KeyStore {
	return &KeyStore{
		rootKey: rootKey,
		persist: persist,
	}
}

// Insert inserts a key into the key store
func (ks *KeyStore) Insert(keyID string, key encryptor.Key) error {
	return ks.persistKey(keyID, key, "DK")
}

// Get retrieves a key from the key store
func (ks *KeyStore) Get(keyID string) (*encryptor.Key, error) {
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

// GetWithTag retrieves a key with a specific tag from the key store
func (ks *KeyStore) GetWithTag(tag string) (string, encryptor.Key, error) {
	sk, err := ks.persist.GetWithTag(tag)
	if err != nil {
		return "", encryptor.Key{}, err
	}

	if sk.ID == "" {
		return "", encryptor.Key{}, nil
	}

	key, err := ks.DeserializeKeyPair(sk.Nonce, sk.Key)
	if err != nil {
		return "", encryptor.Key{}, err
	}

	return sk.ID, *key, nil
}

// persistKey handles the logic of persisting a key
func (ks *KeyStore) persistKey(keyID string, key encryptor.Key, tag string) error {
	// Implement serialization and hashing logic
	nonceHashed, keyHashed, err := ks.SerializeKeyPair(key[:])
	if err != nil {
		return err
	}

	sk := SerializedKey{
		ID:    keyID,
		Nonce: nonceHashed,
		Key:   keyHashed,
		Tag:   tag,
	}

	return ks.persist.SaveKey(sk)
}
