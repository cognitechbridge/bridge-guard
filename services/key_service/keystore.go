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
	KeyNotFound        = errors.New("key not found")
)

type Key = core.Key

// KeyStoreDefault represents a key store
type KeyStoreDefault struct {
	userId          string
	secret          string
	privateKey      []byte
	recoveryItems   []core.RecoveryItem
	keyRepository   repositories.KeyRepository
	vaultRepository repositories.VaultRepository
}

var _ core.KeyService = &KeyStoreDefault{}

// NewKeyStore creates a new instance of KeyStoreDefault
func NewKeyStore(userId string, keyRepository repositories.KeyRepository, vaultRepository repositories.VaultRepository) *KeyStoreDefault {
	return &KeyStoreDefault{
		userId:          userId,
		keyRepository:   keyRepository,
		vaultRepository: vaultRepository,
		recoveryItems:   make([]core.RecoveryItem, 0),
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
func (ks *KeyStoreDefault) Get(keyId string, startVaultId string) (*core.KeyInfo, error) {
	if err := ks.LoadKeys(); err != nil {
		return nil, fmt.Errorf("cannot load keys: %v", err)
	}
	//Check Direct
	if ks.keyRepository.DataKeyExist(keyId, ks.userId) {
		sk, err := ks.keyRepository.GetDataKey(keyId, ks.userId)
		if err != nil {
			return nil, err
		}
		key, err := key_crypto.OpenDataKey(sk, ks.privateKey)
		if err != nil {
			return nil, err
		}
		keyInfo := core.NewKeyInfo(keyId, key[:])
		return &keyInfo, nil
	}
	if startVaultId == "" {
		return nil, KeyNotFound
	}
	vault, err := ks.vaultRepository.GetVault(startVaultId)
	if err != nil {
		return nil, err
	}
	encKey, found := ks.vaultRepository.GetKey(keyId, vault.Id)
	if !found {
		return nil, KeyNotFound
	}
	vaultKey, err := ks.Get(vault.KeyId, vault.ParentId)
	if err != nil {
		return nil, err
	}
	key, err := key_crypto.OpenVaultDataKey(encKey, vaultKey.Key[:])
	if err != nil {
		return nil, err
	}
	keyInfo := core.NewKeyInfo(keyId, key[:])
	return &keyInfo, nil
}

func (ks *KeyStoreDefault) Share(keyId string, recipient []byte, recipientUserId string) error {
	if err := ks.LoadKeys(); err != nil {
		return fmt.Errorf("cannot load keys: %v", err)
	}

	//@Todo: Fix it
	key, err := ks.Get(keyId, "")
	if err != nil {
		return fmt.Errorf("cannot load key: %v", err)
	}

	keyHashed, err := key_crypto.SealDataKey(key.Key[:], recipient)
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

func (ks *KeyStoreDefault) CreateVault(parentId string) (*core.Vault, error) {
	id, err := core.NewUid()
	if err != nil {
		return nil, fmt.Errorf("error generating vault id")
	}
	var key *core.KeyInfo
	if parentId != "" {
		key, err = ks.GenerateKeyInVault(parentId)
		if err != nil {
			return nil, fmt.Errorf("error generating key")
		}
	} else {
		key, err = recovery.GenerateKey(make([]core.RecoveryItem, 0))
		err = ks.Insert(key)
		if err != nil {
			return nil, err
		}
	}
	vault := core.Vault{
		Id:       id,
		KeyId:    key.Id,
		ParentId: parentId,
	}
	err = ks.vaultRepository.InsertVault(vault)
	if err != nil {
		return nil, err
	}
	return &vault, nil
}

func (ks *KeyStoreDefault) AddKeyToVault(vault *core.Vault, key core.KeyInfo) error {
	vKey, err := ks.Get(vault.KeyId, vault.ParentId)
	if err != nil {
		return err
	}
	sealedKey, err := key_crypto.SealVaultDataKey(key.Key, vKey.Key[:])
	if err != nil {
		return err
	}
	err = ks.vaultRepository.AddKeyToVault(vault, key.Id, sealedKey)
	if err != nil {
		return err
	}
	return nil
}

func (ks *KeyStoreDefault) MoveVault(vaultId string, oldParentVaultId string, newParentVaultId string) error {
	vault, err := ks.vaultRepository.GetVault(vaultId)
	if err != nil {
		return err
	}
	err = ks.MoveKey(vault.KeyId, oldParentVaultId, newParentVaultId)
	if err != nil {
		return err
	}
	vault.ParentId = newParentVaultId
	err = ks.vaultRepository.SaveVault(vault)
	if err != nil {
		return err
	}
	return nil
}

func (ks *KeyStoreDefault) MoveKey(keyId string, oldVaultId string, newVaultId string) error {
	key, err := ks.Get(keyId, oldVaultId)
	if err != nil {
		return err
	}

	//Add key to new vault
	newVault, err := ks.vaultRepository.GetVault(newVaultId)
	if err != nil {
		return err
	}
	err = ks.AddKeyToVault(&newVault, *key)
	if err != nil {
		return err
	}

	// Remove key from old vault
	err = ks.vaultRepository.RemoveKey(keyId, oldVaultId)
	if err != nil {
		return err
	}

	return nil
}

func (ks *KeyStoreDefault) GenerateKeyInVault(vaultId string) (*core.KeyInfo, error) {
	vault, err := ks.vaultRepository.GetVault(vaultId)
	if err != nil {
		return nil, err
	}
	key, err := recovery.GenerateKey(make([]core.RecoveryItem, 0))
	if err != nil {
		return nil, err
	}
	err = ks.AddKeyToVault(&vault, *key)
	if err != nil {
		return nil, err
	}
	return key, nil
}
