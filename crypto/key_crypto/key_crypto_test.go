package key_crypto_test

import (
	"bytes"
	"crypto/rand"
	"ctb-cli/core"
	"ctb-cli/crypto/key_crypto"
	"testing"
)

func TestSealAndOpenVaultDataKey(t *testing.T) {
	// Generate a random data key and vault key
	dataKey := make([]byte, 32)
	vaultKey := make([]byte, 32)
	_, err := rand.Read(dataKey)
	if err != nil {
		t.Fatal(err)
	}
	_, err = rand.Read(vaultKey)
	if err != nil {
		t.Fatal(err)
	}

	// Seal the data key
	sealedKey, err := key_crypto.SealVaultDataKey(dataKey, vaultKey)
	if err != nil {
		t.Fatal(err)
	}

	// Open the sealed key
	openedKey, err := key_crypto.OpenVaultDataKey(sealedKey, vaultKey)
	if err != nil {
		t.Fatal(err)
	}

	// Check that the opened key matches the original data key
	if !bytes.Equal(openedKey[:], dataKey) {
		t.Errorf("Opened key does not match original data key")
	}
}

func TestSealAndOpenDataKey(t *testing.T) {
	// Generate a random data key and private key
	dataKey := make([]byte, 32)
	_, err := rand.Read(dataKey)
	if err != nil {
		t.Fatal(err)
	}
	privateKey, err := core.NewPrivateKeyFromRand()
	if err != nil {
		t.Fatal(err)
	}

	publicKey, _ := privateKey.ToPublicKey()

	// Seal the data key
	sealedKey, err := key_crypto.SealDataKey(dataKey, publicKey)
	if err != nil {
		t.Fatal(err)
	}

	// Open the sealed key
	openedKey, err := key_crypto.OpenDataKey(sealedKey, privateKey)
	if err != nil {
		t.Fatal(err)
	}

	// Check that the opened key matches the original data key
	if !bytes.Equal(openedKey[:], dataKey) {
		t.Errorf("Opened key does not match original data key")
	}
}
