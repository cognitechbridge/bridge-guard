package keystore

import (
	"storage-go/encryptor"
	"sync"
)

type SerializedKey struct {
	ID    string
	Nonce string
	Key   string
	Tag   string
}

// KeyStorePersist is an interface for persisting keys
type KeyStorePersist interface {
	SaveKey(serializedKey SerializedKey) error
	GetKey(keyID string) (*SerializedKey, error)
	GetWithTag(tag string) (*SerializedKey, error)
}

// KeyStore represents a key store
type KeyStore struct {
	rootKey encryptor.Key
	persist KeyStorePersist
	mu      sync.RWMutex
}

// NewKeyStore creates a new instance of KeyStore
func NewKeyStore(rootKey encryptor.Key, persist KeyStorePersist) *KeyStore {
	return &KeyStore{
		rootKey: rootKey,
		persist: persist,
	}
}

// Insert inserts a key into the key store
func (ks *KeyStore) Insert(keyID string, key encryptor.Key) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	return ks.persistKey(keyID, key, "DK")
}

// Get retrieves a key from the key store
func (ks *KeyStore) Get(keyID string) (encryptor.Key, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	sk, err := ks.persist.GetKey(keyID)
	if err != nil {
		return encryptor.Key{}, err
	}

	if sk.ID == "" {
		return encryptor.Key{}, nil
	}

	key, err := ks.DeserializeKeyPair(sk.Nonce, sk.Key)
	if err != nil {
		return encryptor.Key{}, err
	}

	return *key, nil
}

// GetWithTag retrieves a key with a specific tag from the key store
func (ks *KeyStore) GetWithTag(tag string) (string, encryptor.Key, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

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
