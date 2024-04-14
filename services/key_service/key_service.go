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
	ErrGeneratingVaultId                = errors.New("error generating vault id")
	ErrGeneratingKey                    = errors.New("error generating key")
)

// KeyStoreDefault represents a key store
type KeyStoreDefault struct {
	privateKey      []byte
	keyRepository   repositories.KeyRepository
	vaultRepository repositories.VaultRepository
}

// Ensure KeyStoreDefault implements KeyService
var _ core.KeyService = &KeyStoreDefault{}

// NewKeyStore creates a new instance of KeyStoreDefault
func NewKeyStore(keyRepository repositories.KeyRepository, vaultRepository repositories.VaultRepository) *KeyStoreDefault {
	return &KeyStoreDefault{
		keyRepository:   keyRepository,
		vaultRepository: vaultRepository,
	}
}

// SetPrivateKey sets the private key in the KeyStoreDefault instance.
func (ks *KeyStoreDefault) SetPrivateKey(privateKey []byte) {
	ks.privateKey = privateKey
}

// GetUserId returns the user ID associated with the key store.
// It retrieves the user's public key and encodes it to obtain the user ID.
func (ks *KeyStoreDefault) GetUserId() (string, error) {
	// Get user public key
	publicKey, err := ks.GetPublicKey()
	if err != nil {
		return "", err
	}
	// Encode public key to get user id
	userId, err := core.EncodePublic(publicKey)
	if err != nil {
		return "", err
	}
	return userId, nil
}

// Insert inserts a new key into the key store.
// It first retrieves the user's public key using the GetPublicKey method.
// Then, it seals the key with the user's public key using the SealDataKey method.
// Next, it retrieves the user's ID using the GetUserId method.
// Finally, it saves the key in the user's data keys using the SaveDataKey method.
// If any error occurs during the process, it is returned.
func (ks *KeyStoreDefault) Insert(key *core.KeyInfo) error {
	// Get user public key
	pk, err := ks.GetPublicKey()
	if err != nil {
		return err
	}
	// Seal key with user public key
	keyHashed, err := key_crypto.SealDataKey(key.Key[:], pk)
	if err != nil {
		return err
	}
	// Get user id
	userId, err := ks.GetUserId()
	if err != nil {
		return err
	}
	// Save key in user's data keys
	return ks.keyRepository.SaveDataKey(key.Id, keyHashed, userId)
}

// Get retrieves a key from the KeyStoreDefault.
// It takes a keyId and a startVaultId as parameters.
// If the key exists in the user's data keys, it returns the key in KeyInfo format.
// If the key does not exist in the user's data keys, it checks if it exists in the provided vault.
// If the key is found in the vault, it recursively calls the Get method to retrieve the vault key.
// It then retrieves the encrypted data key from the vault and unseals it using the vault key.
// Finally, it returns the key in KeyInfo format.
func (ks *KeyStoreDefault) Get(keyId string, startVaultId string) (*core.KeyInfo, error) {
	// Get user id
	userId, err := ks.GetUserId()
	if err != nil {
		return nil, err
	}
	// Check if key directly exists in user's data keys
	if ks.keyRepository.DataKeyExist(keyId, userId) {
		// Get key from user's data keys
		sk, err := ks.keyRepository.GetDataKey(keyId, userId)
		if err != nil {
			return nil, err
		}
		// Unseal key
		key, err := key_crypto.OpenDataKey(sk, ks.privateKey)
		if err != nil {
			return nil, err
		}
		// Return key in KeyInfo format
		keyInfo := core.NewKeyInfo(keyId, key[:])
		return &keyInfo, nil
	}
	// If key does not exist in user's data keys, check if it exists in a vault
	// If startVaultId is not provided, return key not found
	if startVaultId == "" {
		return nil, ErrDataKeyNotFound

	}
	// Get start vault
	vault, err := ks.vaultRepository.GetVault(startVaultId)
	if err != nil {
		return nil, err
	}
	//Get encrypted data key from vault
	encKey, found := ks.vaultRepository.GetKey(keyId, vault.Id)
	if !found {
		return nil, ErrDataKeyNotFound
	}
	// Get vault key using recursive call
	vaultKey, err := ks.Get(vault.KeyId, vault.ParentId)
	if err != nil {
		return nil, err
	}
	// Unseal key using vault key
	key, err := key_crypto.OpenVaultDataKey(encKey, vaultKey.Key[:])
	if err != nil {
		return nil, err
	}
	// Return key in KeyInfo format
	keyInfo := core.NewKeyInfo(keyId, key[:])
	return &keyInfo, nil
}

// GetHasAccessToKey checks if a user has access to a specific key.
// It also checks if the access is inherited from a vault or directly from the user's data keys.
// It first checks if the key directly exists in the user's data keys.
// If the key exists, it returns true and false for `hasAccess` and `inherited` respectively.
// If the key does not exist in the user's data keys, it checks if it exists in a vault.
// If the key exists in a vault, it recursively calls `GetHasAccessToKey` to check if the user has access to the vault key.
// It returns the result of the recursive call and true for `inherited`.
func (ks *KeyStoreDefault) GetHasAccessToKey(keyId string, startVaultId string, userId string) (hasAccess bool, inherited bool) {
	// Check if key directly exists in user's data keys
	if ks.keyRepository.DataKeyExist(keyId, userId) {
		// Get key from user's data keys
		exc := ks.keyRepository.DataKeyExist(keyId, userId)
		if exc == true {
			return true, false
		}
	}
	// If key does not exist in user's data keys, check if it exists in a vault
	// If startVaultId is not provided, return false
	if startVaultId == "" {
		return false, false

	}
	// Get start vault
	vault, err := ks.vaultRepository.GetVault(startVaultId)
	if err != nil {
		return false, false
	}
	// Get vault key using recursive call to GetHasAccessToKey
	px, _ := ks.GetHasAccessToKey(vault.KeyId, vault.ParentId, userId)
	return px, true
}

func (ks *KeyStoreDefault) Share(keyId string, startVaultId string, recipient []byte, recipientUserId string) error {
	key, err := ks.Get(keyId, startVaultId)
	if err != nil {
		return fmt.Errorf("cannot load key: %v", err)
	}

	keyHashed, err := key_crypto.SealDataKey(key.Key[:], recipient)
	if err != nil {
		return err
	}

	return ks.keyRepository.SaveDataKey(keyId, keyHashed, recipientUserId)
}

// GetPublicKey returns the public key corresponding to the private key stored in the KeyStore.
// It uses the X25519 function from the curve25519 package to perform the scalar multiplication
// of the private key with the base point, resulting in the public key.
// If any error occurs during the process, it returns the error.
func (ks *KeyStoreDefault) GetPublicKey() ([]byte, error) {
	return curve25519.X25519(ks.privateKey, curve25519.Basepoint)
}

// GetEncodablePublicKey returns the public key as a string.
// It encode the public key using base58 encoding and returns it.
// If any error occurs during the process, it returns the error.
func (ks *KeyStoreDefault) GetEncodablePublicKey() (string, error) {
	publicKey, err := ks.GetPublicKey()
	if err != nil {
		return "", err
	}
	return core.EncodePublic(publicKey)
}

// CreateVault generates a new vault and inserts it into the vault repository.
// If parentId is provided, it generates a key in the parent vault and associates it with the new vault.
// If parentId is not provided, it generates a key without a parent and inserts it into the keystore.
// The generated vault and associated key are returned on success.
// If any error occurs during the process, an error is returned.
func (ks *KeyStoreDefault) CreateVault(parentId string) (*core.Vault, error) {
	// Generate vault id
	id, err := core.NewUid()
	if err != nil {
		return nil, ErrGeneratingVaultId
	}
	// Generate vault key
	var key *core.KeyInfo
	if parentId != "" {
		// If parentId is not empty, generate key in parent vault
		key, err = ks.GenerateKeyInVault(parentId)
		if err != nil {
			return nil, ErrGeneratingKey
		}
	} else {
		// If parentId is empty, generate key without parent
		key, err = core.GenerateKey()
		if err != nil {
			return nil, ErrGeneratingKey
		}
		// Insert key into keystore
		err = ks.Insert(key)
		if err != nil {
			return nil, err
		}
	}
	// Insert vault into vault repository
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

// AddKeyToVault adds a key to the specified vault.
// It retrieves the vault key, seals the provided key with the vault key,
// and adds the sealed key to the vault.
// If any error occurs during the process, it returns the error.
func (ks *KeyStoreDefault) AddKeyToVault(vault *core.Vault, key core.KeyInfo) error {
	// Get vault key
	vKey, err := ks.Get(vault.KeyId, vault.ParentId)
	if err != nil {
		return err
	}
	// Seal key with vault key
	sealedKey, err := key_crypto.SealVaultDataKey(key.Key, vKey.Key[:])
	if err != nil {
		return err
	}
	// Add sealed key to vault
	err = ks.vaultRepository.AddKeyToVault(vault, key.Id, sealedKey)
	if err != nil {
		return err
	}
	return nil
}

// MoveVault moves a vault to a new parent vault.
// It first retrieves the vault using the provided vaultId.
// Then, it moves the vault key to the new parent vault using the MoveKey function.
// Finally, it updates the vault's parent and saves the changes using the vaultRepository.
// If any error occurs during the process, it is returned.
func (ks *KeyStoreDefault) MoveVault(vaultId string, oldParentVaultId string, newParentVaultId string) error {
	// Get vault to find vault key id
	vault, err := ks.vaultRepository.GetVault(vaultId)
	if err != nil {
		return err
	}
	// Move vault key to new parent
	err = ks.MoveKey(vault.KeyId, oldParentVaultId, newParentVaultId)
	if err != nil {
		return err
	}
	// Update vault parent
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
	if !ks.keyRepository.IsUserJoined(userId) {
		return false, ErrInvalidPrivateKeyOrUserNotJoined
	}
	return true, nil
}

// Join joins the user to a group by retrieving the user ID and calling JoinByUserId.
// It returns an error if there was an issue retrieving the user ID or joining the group.
func (ks *KeyStoreDefault) Join() error {
	userId, err := ks.GetUserId()
	if err != nil {
		return err
	}
	return ks.JoinByUserId(userId)
}

// JoinByUserId joins a user by their user ID.
// It checks if the user is already joined and returns an error if so.
// Otherwise, it calls the key repository to join the user and returns any error that occurs.
func (ks *KeyStoreDefault) JoinByUserId(userId string) error {
	if ks.keyRepository.IsUserJoined(userId) {
		return ErrUserAlreadyJoined
	}
	err := ks.keyRepository.JoinUser(userId)
	if err != nil {
		return err
	}
	return nil
}

// GenerateUserKey generates a new user key and returns it as a string.
// If any error occurs during the process, it returns the error.
func (ks *KeyStoreDefault) GenerateUserKey() (*core.UserKeyPair, error) {
	key, err := core.GenerateUserKey()
	if err != nil {
		return nil, err
	}
	return key, nil
}

// IsUserJoined checks if the user with the specified user ID has joined the key store.
// It returns a boolean indicating whether the user has joined or not.
// If any error occurs during the process, it returns false.
func (ks *KeyStoreDefault) IsUserJoined() bool {
	userId, err := ks.GetUserId()
	if err != nil {
		return false
	}
	return ks.keyRepository.IsUserJoined(userId)
}

// GetKeyAccessList retrieves the key access list for a given key ID and starting vault ID.
// It returns a list of KeyAccess objects representing the users who have access to the key,
// along with a boolean value indicating whether the access is inherited from a parent vault.
// If an error occurs during the retrieval process, it is returned as the second value.
func (ks *KeyStoreDefault) GetKeyAccessList(keyId string, startVaultId string) (core.KeyAccessList, error) {
	usersList, err := ks.keyRepository.ListUsers()
	if err != nil {
		return nil, err
	}
	accessList := make(core.KeyAccessList, 0)
	for _, user := range usersList {
		if hasAccess, inherited := ks.GetHasAccessToKey(keyId, startVaultId, user); hasAccess {
			accessList = append(accessList, core.KeyAccess{
				PublicKey: user,
				Inherited: inherited,
			})
		}
	}
	return accessList, nil
}
