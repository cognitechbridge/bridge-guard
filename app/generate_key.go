package app

import (
	"crypto/rand"
	"ctb-cli/config"
	"ctb-cli/core"
	"io"
)

func GenerateKey(secret string, email string) error {
	//Generate random use id
	bytes := make([]byte, 128/8)
	_, err := io.ReadFull(rand.Reader, bytes)
	if err != nil {
		return err
	}
	userId, err := core.EncodeUid(bytes)
	if err != nil {
		return err
	}

	//Save user id to config and keystore
	err = config.Workspace.SetUserId(userId)
	keyStore.SetUserId(userId)
	if err != nil {
		return err
	}

	keyStore.SetSecret(secret)
	if err != nil {
		return err
	}
	err = keyStore.GenerateUserKeys()
	if err != nil {
		return err
	}

	publicKey, err := keyStore.GetPublicKey()
	if err != nil {
		return err
	}
	recipient, err := core.NewRecipient(email, publicKey, userId)
	if err != nil {
		return err
	}
	err = shareService.SaveRecipient(recipient)
	if err != nil {
		return err
	}
	return nil
}
