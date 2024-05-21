package key_crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"ctb-cli/core"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

const (
	X25519V1Info           = "cognitechbridge.com/v1/X25519"           // X25519V1Info is the info string used for deriving the wrap key from the shared secret.
	ChaCha20Poly1350V1Info = "cognitechbridge.com/v1/ChaCha20Poly1350" // ChaCha20Poly1350V1Info is the info string used for deriving the encryption key from the vault key.
)

var (
	ErrGeneratingRandomSalt            = errors.New("error generating random salt")
	ErrInvalidKey                      = errors.New("invalid key")
	ErrGeneratingDerivedKey            = errors.New("error generating derived key")
	ErrInvalidSerializedKey            = errors.New("invalid serialized key")
	ErrGeneratingRandomEphemeralSecret = errors.New("error generating random ephemeral secret")
	ErrFaliledToCreateCipher           = errors.New("failed to create cipher")
	ErrErrorDerivingWrapKey            = errors.New("error deriving wrap key")
)

// deriveKey derives a key from the root key, salt, and info using HKDF and SHA-256.
// It returns the derived key and any error encountered during the derivation process.
func deriveKey(rootKey []byte, salt []byte, info string) (core.Key, error) {
	// Derive a key from the root key, salt, and info using HKDF and SHA-256
	hk := hkdf.New(sha256.New, rootKey[:], salt, []byte(info))
	derivedKey := core.Key{}
	_, err := io.ReadFull(hk, derivedKey[:])
	return derivedKey, err
}

// SealVaultDataKey encrypts the given data key using a vault key and returns the encrypted result.
// It generates a random 32-byte salt, derives a key from the vault key, salt, and info using HKDF and SHA-256,
// creates a new AEAD cipher using the derived key, encrypts the data key using the AEAD cipher,
// and serializes the salt and ciphered data key.
// The result is returned as a string in the format "salt:cipheredDataKey".
// If any error occurs during the process, an error is returned.
func SealVaultDataKey(dataKey []byte, vaultKey []byte) (string, error) {
	// Generate a random 32-byte salt
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		return "", ErrGeneratingRandomSalt
	}
	// Derive a key from the vault key, salt, and info using HKDF and SHA-256
	derivedKey, err := deriveKey(vaultKey, salt, ChaCha20Poly1350V1Info)
	if err != nil {
		return "", ErrGeneratingDerivedKey
	}
	// Create a new AEAD cipher using the derived key
	aead, err := chacha20poly1305.New(derivedKey[:])
	if err != nil {
		return "", ErrFaliledToCreateCipher
	}
	// Create a all-zero nonce
	nonce := make([]byte, chacha20poly1305.NonceSize)
	// Encrypt the data key using the AEAD cipher
	ciphered := aead.Seal(nil, nonce, dataKey, nil)
	// Serialize the salt and ciphered data key
	res := fmt.Sprintf("%s:%s",
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(ciphered),
	)

	return res, nil
}

// OpenVaultDataKey decrypts a serialized key using a vault key.
// It splits the serialized key into the salt and ciphered data key,
// decodes them from raw base64, and retrieves the derived key from
// the vault key, salt, and info using HKDF and SHA-256. Then, it
// creates an AEAD cipher using the derived key and decrypts the data
// key using the AEAD cipher. Finally, it converts the deciphered data
// key to a core.Key format and returns it.
//
// Parameters:
//   - serialized: The serialized key to be decrypted.
//   - vaultKey: The vault key used to derive the encryption key.
//
// Returns:
//   - *core.Key: The decrypted data key.
//   - error: An error if decryption fails or the serialized key is invalid.
func OpenVaultDataKey(serialized string, vaultKey []byte) (*core.Key, error) {
	// Split the serialized key into the salt and ciphered data key by the colon separator
	parts := strings.Split(serialized, ":")
	if len(parts) != 2 {
		return nil, ErrInvalidSerializedKey
	}
	// Decode the salt and ciphered data key from raw base64
	salt, err1 := base64.RawStdEncoding.DecodeString(parts[0])
	ciphered, err2 := base64.RawStdEncoding.DecodeString(parts[1])
	if errors.Join(err1, err2) != nil {
		return nil, ErrInvalidSerializedKey
	}
	// Retrieve the derived key from the vault key, salt, and info using HKDF and SHA-256
	derivedKey, err := deriveKey(vaultKey, salt, ChaCha20Poly1350V1Info)
	if err != nil {
		return nil, ErrGeneratingDerivedKey
	}
	// Create AEAD cipher using the derived key
	aead, err := chacha20poly1305.New(derivedKey[:])
	if err != nil {
		return nil, ErrFaliledToCreateCipher
	}
	// Create a all-zero nonce
	nonce := make([]byte, chacha20poly1305.NonceSize)
	// Decrypt the data key using the AEAD cipher
	deciphered, err := aead.Open(nil, nonce, ciphered, nil)
	if err != nil {
		return nil, ErrInvalidKey
	}
	// Convert the deciphered data key to a core.Key format
	key := core.Key{}
	copy(key[:], deciphered)
	return &key, nil
}

// SealDataKey encrypts a data key using public key encryption and returns the encrypted result.
// It takes a key byte slice and a public key byte slice as input parameters.
// The function generates a random 32-byte ephemeral secret, derives the ephemeral share from the ephemeral secret and the basepoint using X25519,
// encodes the ephemeral share and public key to raw base64, derives the shared secret from the ephemeral secret and the public key using X25519,
// derives the wrap key from the shared secret, salt, and info using HKDF and SHA-256, creates a new AEAD cipher using the wrap key,
// encrypts the data key using the AEAD cipher, and serializes the ephemeral share and ciphered data key.
// The function returns the serialized result as a string in the format "ephemeralShare \n cipheredDataKey" and any error encountered during the encryption process.
func SealDataKey(key []byte, publicKey core.PublicKey) (string, error) {
	// Generate a random 32-byte ephemeral secret
	ephemeralSecret := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, ephemeralSecret[:])
	if err != nil {
		return "", ErrGeneratingRandomEphemeralSecret
	}
	// Derive the ephemeral share from the ephemeral secret and the basepoint using X25519
	ephemeralShare, err := curve25519.X25519(ephemeralSecret, curve25519.Basepoint)
	if err != nil {
		return "", fmt.Errorf("error encrypting data key: %v", err)
	}
	// Encode the ephemeral share and public key to raw base64
	ephemeralShareString := base64.RawStdEncoding.EncodeToString(ephemeralShare)
	// Generate the salt from the ephemeral share and encoded public key
	salt := ephemeralShareString + publicKey.Encode()
	// Derive the shared secret from the ephemeral secret and the public key using X25519
	sharedSecret, err := curve25519.X25519(ephemeralSecret, publicKey.Bytes())
	if err != nil {
		return "", fmt.Errorf("error encrypting data key: %v", err)
	}
	// Derive the wrap key from the shared secret, salt, and info using HKDF and SHA-256
	wrapKey, err := deriveKey(sharedSecret, []byte(salt), X25519V1Info)
	if err != nil {
		return "", fmt.Errorf("error generating wrap key: %v", err)
	}
	// Create a new AEAD cipher using the wrap key
	aead, err := chacha20poly1305.New(wrapKey[:])
	if err != nil {
		return "", ErrFaliledToCreateCipher
	}
	// Create a all-zero nonce
	nonce := make([]byte, chacha20poly1305.NonceSize)
	// Encrypt the data key using the AEAD cipher
	ciphered := aead.Seal(nil, nonce, key, nil)
	// Serialize the ephemeral share and ciphered data key
	res := fmt.Sprintf("%s\n%s",
		ephemeralShareString,
		base64.RawStdEncoding.EncodeToString(ciphered),
	)
	return res, nil
}

// OpenDataKey decrypts a serialized key using the provided private key.
// It splits the serialized key into the ephemeral share and ciphered data key,
// decodes them from base64, and derives the shared secret and wrap key.
// Finally, it decrypts the data key using the wrap key and returns it as a core.Key.
//
// Parameters:
//   - serialized: The serialized key to be decrypted.
//   - privateKey: The private key used for decryption.
//
// Returns:
//   - *core.Key: The decrypted data key.
//   - error: An error if decryption fails.
func OpenDataKey(serialized string, privateKey core.PrivateKey) (*core.Key, error) {
	// Split the serialized key into the ephemeral share and ciphered data key by the newline separator
	parts := strings.Split(serialized, "\n")
	if len(parts) != 2 {
		return nil, ErrInvalidSerializedKey
	}
	// Decode the ephemeral share and ciphered data key from raw base64
	ephemeralShareString := parts[0]
	ephemeralShare, err1 := base64.RawStdEncoding.DecodeString(ephemeralShareString)
	ciphered, err2 := base64.RawStdEncoding.DecodeString(parts[1])
	if errors.Join(err1, err2) != nil {
		return nil, ErrInvalidSerializedKey
	}
	// Derive the public key from the private key using X25519
	publicKey, err := privateKey.ToPublicKey()
	if err != nil {
		return nil, fmt.Errorf("error decrypting data key: %v", err)
	}
	// Regenerate the salt from the ephemeral share and encoded public key
	salt := ephemeralShareString + publicKey.Encode()
	// Derive the shared secret from the ephemeral share and the private key using X25519
	sharedSecret, err := curve25519.X25519(privateKey.Bytes(), ephemeralShare)
	if err != nil {
		return nil, fmt.Errorf("error decrypting data key: %v", err)
	}
	// Derive the wrap key from the shared secret, salt, and info using HKDF and SHA-256
	wrapKey, err := deriveKey(sharedSecret, []byte(salt), X25519V1Info)
	if err != nil {
		return nil, ErrErrorDerivingWrapKey
	}
	// Create a new AEAD cipher using the wrap key
	aead, err := chacha20poly1305.New(wrapKey[:])
	if err != nil {
		return nil, ErrFaliledToCreateCipher
	}
	// Create a all-zero nonce
	nonce := make([]byte, chacha20poly1305.NonceSize)
	// Decrypt the data key using the AEAD cipher
	deciphered, err := aead.Open(nil, nonce, ciphered, nil)
	if err != nil {
		return nil, fmt.Errorf("error decrypting data key: %v", err)
	}
	// Convert the deciphered data key to a core.Key format
	key := core.Key{}
	copy(key[:], deciphered)

	return &key, nil
}
