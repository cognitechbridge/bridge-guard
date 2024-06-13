package app

import (
	"ctb-cli/core"
	"ctb-cli/services/key_service"
)

// GenerateUserKey generates a user private key and returns it as a string.
// It returns an AppResult containing the generated key on success,
// or an AppErrorResult containing the error on failure.
func (a *App) GenerateUserKey() core.AppResult {
	keyStore := key_service.NewKeyStore(nil, nil)
	// generate the key
	key, err := keyStore.GenerateUserKey()
	if err != nil {
		return core.NewAppResultWithError(err)
	}
	publicKey, err := key.ToPublicKey()
	if err != nil {
		return core.NewAppResultWithError(err)
	}
	return core.NewAppResultWithValue(GenerateUserKeyResult{
		PrivateKey: key.Unsafe().String(),
		PublicKey:  publicKey.String(),
	})
}

type GenerateUserKeyResult struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}
