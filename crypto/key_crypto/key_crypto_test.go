package key_crypto_test

import (
	"ctb-cli/core"
	"ctb-cli/crypto/key_crypto"
	"testing"
)

func TestSealAndOpenVaultDataKey(t *testing.T) {
	// Generate a random data key and vault key
	dataKey := core.NewKeyFromRand()
	vaultKey := core.NewKeyFromRand()

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
	if !openedKey.Equals(dataKey) {
		t.Errorf("Opened key does not match original data key")
	}
}

func TestSealAndOpenDataKey(t *testing.T) {
	// Generate a random data key and private key
	dataKey := core.NewKeyFromRand()
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
	if !openedKey.Equals(dataKey) {
		t.Errorf("Opened key does not match original data key")
	}
}
