package app

import (
	"crypto/rand"
	"ctb-cli/config"
	"ctb-cli/core"
	"io"
)

func GenerateKey(secret string, email string) core.AppResult {
	//Generate random use id
	bytes := make([]byte, 128/8)
	_, err := io.ReadFull(rand.Reader, bytes)
	if err != nil {
		return core.AppErrorResult(err)
	}
	userId, err := core.EncodeUid(bytes)
	if err != nil {
		return core.AppErrorResult(err)
	}

	//Save user id to config and keystore
	err = config.Workspace.SetUserId(userId)
	keyStore.SetUserId(userId)
	if err != nil {
		return core.AppErrorResult(err)
	}

	keyStore.SetSecret(secret)
	if err != nil {
		return core.AppErrorResult(err)
	}
	err = keyStore.GenerateUserKeys()
	if err != nil {
		return core.AppErrorResult(err)
	}

	publicKey, err := keyStore.GetPublicKey()
	if err != nil {
		return core.AppErrorResult(err)
	}
	recipient, err := core.NewRecipient(email, publicKey, userId)
	if err != nil {
		return core.AppErrorResult(err)
	}
	err = shareService.SaveRecipient(recipient)
	if err != nil {
		return core.AppErrorResult(err)
	}
	return core.AppOkResult()
}
