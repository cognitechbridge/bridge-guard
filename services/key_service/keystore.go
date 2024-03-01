package key_service

import (
	"ctb-cli/core"
	"ctb-cli/crypto/key_crypto"
	"ctb-cli/crypto/recovery"
	"ctb-cli/repositories"
	"errors"
	"fmt"
	"golang.org/x/crypto/curve25519"
	"os"
)

var (
	ErrorInvalidSecret = errors.New("invalid secret")
)

type Key = core.Key

// KeyStoreDefault represents a key store
type KeyStoreDefault struct {
	userId        string
	secret        string
	privateKey    []byte
	recoveryItems []core.RecoveryItem
	keyRepository repositories.KeyRepository
}

var _ core.KeyService = &KeyStoreDefault{}

// NewKeyStore creates a new instance of KeyStoreDefault
func NewKeyStore(userId string, keyRepository repositories.KeyRepository) *KeyStoreDefault {
	return &KeyStoreDefault{
		userId:        userId,
		keyRepository: keyRepository,
		recoveryItems: make([]core.RecoveryItem, 0),
	}
}

func (ks *KeyStoreDefault) SetUserId(userId string) {
	ks.userId = userId
}

// Insert inserts a key into the key store
func (ks *KeyStoreDefault) Insert(key *core.KeyInfo) error {
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
	sk, err := ks.keyRepository.GetDataKey(keyID, ks.userId)
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

func (ks *KeyStoreDefault) GetRecoveryItems() ([]core.RecoveryItem, error) {
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
	serializedPrivateKey, err := ks.keyRepository.GetPrivateKey(ks.userId)
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
	privateKey := core.NewKeyFromRand()
	ks.privateKey = privateKey[:]
	//Save private key
	sealPrivateKey, err := key_crypto.SealPrivateKey(privateKey[:], ks.secret)
	if err != nil {
		return err
	}
	err = ks.keyRepository.SavePrivateKey(sealPrivateKey, ks.userId)
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
	err = ks.keyRepository.SavePrivateKey(sealPrivateKey, ks.userId)
	if err != nil {
		return err
	}
	return nil
}

func (ks *KeyStoreDefault) CreateVault() (*core.Vault, error) {
	id, err := core.NewUid()
	if err != nil {
		return nil, fmt.Errorf("error generating vault id")
	}
	key, err := recovery.GenerateKey(make([]core.RecoveryItem, 0))
	if err != nil {
		return nil, fmt.Errorf("error generating key")
	}
	err = ks.Insert(key)
	if err != nil {
		return nil, err
	}
	vault := core.Vault{
		Id:            id,
		KeyId:         key.Id,
		EncryptedKeys: make(map[string]string),
	}
	err = ks.keyRepository.SaveVault(vault)
	if err != nil {
		return nil, err
	}
	return &vault, nil
}

func (ks *KeyStoreDefault) GenerateKeyInVault(vaultId string) (*core.KeyInfo, error) {
	vault, err := ks.keyRepository.GetVault(vaultId)
	if err != nil {
		return nil, err
	}
	vKey, err := ks.Get(vault.KeyId)
	if err != nil {
		return nil, err
	}
	key, err := recovery.GenerateKey(make([]core.RecoveryItem, 0))
	if err != nil {
		return nil, err
	}
	sealedKey, err := key_crypto.SealVaultDataKey(key.Key, vKey[:])
	if err != nil {
		return nil, err
	}
	err = vault.AddKey(sealedKey, key.Id)
	if err != nil {
		return nil, err
	}
	err = ks.keyRepository.SaveVault(vault)
	if err != nil {
		return nil, err
	}
	return key, nil
}
