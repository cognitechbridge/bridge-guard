package key_service

import (
	"ctb-cli/core"
	"ctb-cli/crypto/key_crypto"
	"ctb-cli/repositories"
	"errors"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

var (
	ErrInvalidPrivateKeyOrUserNotJoined = errors.New("invalid private key or user not joined")
	ErrDataKeyNotFound                  = errors.New("data key not found")
	ErrUserAlreadyJoined                = errors.New("user already joined")
)

type Key = core.Key

// KeyStoreDefault represents a key store
type KeyStoreDefault struct {
	privateKey      []byte
	keyRepository   repositories.KeyRepository
	vaultRepository repositories.VaultRepository
}

// Ensure KeyStoreDefault implements KeyService
var _ core.KeyService = &KeyStoreDefault{}

// NewKeyStore creates a new instance of KeyStoreDefault
func NewKeyStore(userId string, keyRepository repositories.KeyRepository, vaultRepository repositories.VaultRepository) *KeyStoreDefault {
	return &KeyStoreDefault{
		keyRepository:   keyRepository,
		vaultRepository: vaultRepository,
	}
}

func (ks *KeyStoreDefault) SetPrivateKey(privateKey []byte) {
	ks.privateKey = privateKey
}

func (ks *KeyStoreDefault) GetUserId() (string, error) {
	publicKey, err := ks.GetPublicKey()
	if err != nil {
		return "", err
	}
	userId, err := core.EncodePublic(publicKey)
	if err != nil {
		return "", err
	}
	return userId, nil
}

// Insert inserts a key into the key store
func (ks *KeyStoreDefault) Insert(key *core.KeyInfo) error {
	pk, err := ks.GetPublicKey()
	if err != nil {
		return err
	}
	keyHashed, err := key_crypto.SealDataKey(key.Key[:], pk)
	if err != nil {
		return err
	}
	userId, err := ks.GetUserId()
	if err != nil {
		return err
	}
	return ks.keyRepository.SaveDataKey(key.Id, keyHashed, userId)
}

// Get retrieves a key from the key store
func (ks *KeyStoreDefault) Get(keyId string, startVaultId string) (*core.KeyInfo, error) {
	userId, err := ks.GetUserId()
	if err != nil {
		return nil, err
	}
	//Check Direct
	if ks.keyRepository.DataKeyExist(keyId, userId) {
		sk, err := ks.keyRepository.GetDataKey(keyId, userId)
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
		return nil, ErrDataKeyNotFound

	}
	vault, err := ks.vaultRepository.GetVault(startVaultId)
	if err != nil {
		return nil, err
	}
	encKey, found := ks.vaultRepository.GetKey(keyId, vault.Id)
	if !found {
		return nil, ErrDataKeyNotFound

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

func (ks *KeyStoreDefault) GetPublicKey() ([]byte, error) {
	return curve25519.X25519(ks.privateKey, curve25519.Basepoint)
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
		key, err = core.GenerateKey()
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

// MoveKey moves a key from one vault to another.
// It retrieves the key from the old vault, adds it to the new vault, and removes it from the old vault.
// If any error occurs during the process, it returns the error.
func (ks *KeyStoreDefault) MoveKey(keyId string, oldVaultId string, newVaultId string) error {
	key, err := ks.Get(keyId, oldVaultId)
	if err != nil {
		return err
	}

	// Add key to new vault
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

// GenerateKeyInVault generates a new key and adds it to the specified vault.
// It returns the generated key information or an error if the operation fails.
func (ks *KeyStoreDefault) GenerateKeyInVault(vaultId string) (*core.KeyInfo, error) {
	vault, err := ks.vaultRepository.GetVault(vaultId)
	if err != nil {
		return nil, err
	}
	key, err := core.GenerateKey()
	if err != nil {
		return nil, err
	}
	err = ks.AddKeyToVault(&vault, *key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// CheckPrivateKey checks if the private key is valid and if the user has joined the repository.
// It returns a boolean indicating whether the private key is valid or not,
// and an error if any occurred during the check.
func (ks *KeyStoreDefault) CheckPrivateKey() (bool, error) {
	userId, err := ks.GetUserId()
	if err != nil {
		return false, err
	}
	if !ks.keyRepository.DataKeyExist(userId, userId) {
		return false, ErrInvalidPrivateKeyOrUserNotJoined
	}
	_, err = ks.keyRepository.GetDataKey(userId, userId)
	if err != nil {
		return false, ErrInvalidPrivateKeyOrUserNotJoined
	}
	return true, nil
}

// Join adds the current user to the key store.
// It first retrieves the user ID using the GetUserId method.
// Then it checks if the user has already joined the key store.
// If the user has already joined, it returns an error.
// If any error occurs during the process, it returns the error.
// If the user is successfully added, it returns nil.
func (ks *KeyStoreDefault) Join() error {
	userId, err := ks.GetUserId()
	if err != nil {
		return err
	}
	// Check if user already joined
	if ks.keyRepository.IsUserJoined(userId) {
		return ErrUserAlreadyJoined
	}
	// Join user
	err = ks.keyRepository.JoinUser(userId)
	if err != nil {
		return err
	}
	return nil
}
