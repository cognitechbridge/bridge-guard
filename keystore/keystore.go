package keystore

import (
	"ctb-cli/crypto/key_crypto"
	"ctb-cli/crypto/recovery"
	"ctb-cli/repositories"
	"ctb-cli/types"
	"errors"
	"fmt"
	"golang.org/x/crypto/curve25519"
	"os"
)

type KeyStorer interface {
	Get(keyID string) (*types.Key, error)
	Insert(key *types.KeyInfo) error
	GetRecoveryItems() ([]types.RecoveryItem, error)
	AddRecoveryKey(inPath string) error
	GenerateUserKeys() (err error)
	SetSecret(secret string)
	LoadKeys() error
	ChangeSecret(secret string) error
	Share(keyId string, recipient []byte, recipientUserId string) error
	GetPublicKey() ([]byte, error)
	SetUserId(userId string)
}

var (
	ErrorInvalidSecret = errors.New("invalid secret")
)

type Key = types.Key

// KeyStoreDefault represents a key store
type KeyStoreDefault struct {
	userId        string
	secret        string
	privateKey    []byte
	recoveryItems []types.RecoveryItem
	keyRepository repositories.KeyRepository
}

var _ KeyStorer = &KeyStoreDefault{}

// NewKeyStore creates a new instance of KeyStoreDefault
func NewKeyStore(userId string, keyRepository repositories.KeyRepository) *KeyStoreDefault {
	return &KeyStoreDefault{
		userId:        userId,
		keyRepository: keyRepository,
		recoveryItems: make([]types.RecoveryItem, 0),
	}
}

func (ks *KeyStoreDefault) SetUserId(userId string) {
	ks.userId = userId
	ks.keyRepository.SetUserId(userId)
}

// Insert inserts a key into the key store
func (ks *KeyStoreDefault) Insert(key *types.KeyInfo) error {
	if err := ks.LoadKeys(); err != nil {
		return fmt.Errorf("cannot load keys: %v", err)
	}

	pk, err := ks.GetPublicKey()
	if err != nil {
		return err
	}
	keyHashed, err := key_crypto.SealDataKey(key.Key[:], pk)
	if err != nil {
		return err
	}

	return ks.keyRepository.SaveDataKey(key.Id, keyHashed, ks.userId)
}

// Get retrieves a key from the key store
func (ks *KeyStoreDefault) Get(keyID string) (*Key, error) {
	if err := ks.LoadKeys(); err != nil {
		return nil, fmt.Errorf("cannot load keys: %v", err)
	}
	sk, err := ks.keyRepository.GetDataKey(keyID)
	if err != nil {
		return nil, err
	}
	key, err := key_crypto.OpenDataKey(sk, ks.privateKey)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (ks *KeyStoreDefault) Share(keyId string, recipient []byte, recipientUserId string) error {
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

	return ks.keyRepository.SaveDataKey(keyId, keyHashed, recipientUserId)
}

func (ks *KeyStoreDefault) GetRecoveryItems() ([]types.RecoveryItem, error) {
	return ks.recoveryItems, nil
}

func (ks *KeyStoreDefault) AddRecoveryKey(inPath string) error {
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

func (ks *KeyStoreDefault) LoadKeys() error {

	if ks.privateKey != nil {
		return nil
	}
	serializedPrivateKey, err := ks.keyRepository.GetPrivateKey()
	if err != nil {
		return err
	}
	ks.privateKey, err = key_crypto.OpenPrivateKey(serializedPrivateKey, ks.secret)
	if errors.Is(err, key_crypto.ErrorInvalidKey) {
		return ErrorInvalidSecret
	} else if err != nil {
		return err
	}
	return nil
}

func (ks *KeyStoreDefault) GenerateUserKeys() (err error) {
	//Generate private key
	privateKey := types.NewKeyFromRand()
	ks.privateKey = privateKey[:]
	//Save private key
	sealPrivateKey, err := key_crypto.SealPrivateKey(privateKey[:], ks.secret)
	if err != nil {
		return err
	}
	err = ks.keyRepository.SavePrivateKey(sealPrivateKey)
	if err != nil {
		return err
	}

	err = ks.LoadKeys()
	return
}

func (ks *KeyStoreDefault) GetPublicKey() ([]byte, error) {
	return curve25519.X25519(ks.privateKey, curve25519.Basepoint)
}

func (ks *KeyStoreDefault) SetSecret(secret string) {
	ks.secret = secret
	return
}

func (ks *KeyStoreDefault) ChangeSecret(secret string) error {
	if err := ks.LoadKeys(); err != nil {
		return fmt.Errorf("cannot load keys: %v", err)
	}
	sealPrivateKey, err := key_crypto.SealPrivateKey(ks.privateKey, secret)
	ks.secret = secret
	err = ks.keyRepository.SavePrivateKey(sealPrivateKey)
	if err != nil {
		return err
	}
	return nil
}
