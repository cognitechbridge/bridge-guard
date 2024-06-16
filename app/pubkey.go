package app

import (
	"ctb-cli/core"
	"ctb-cli/services/key_service"
)

// GetPubkey generates a public key from a private key.
func (a *App) GetPubkey(privateKeyString string) core.AppResult {
	privateKey, err := core.NewPrivateKeyFromEncoded(privateKeyString)
	if err != nil {
		return core.NewAppResultWithError(err)
	}
	// Create a new key store without key and vault repositories.
	keyStore := key_service.NewKeyStore(nil, nil)
	publicKey, err := keyStore.GetPublicKeyByPrivateKey(privateKey)
	if err != nil {
		return core.NewAppResultWithError(err)
	}

	return core.NewAppResultWithValue(publicKey)
}
