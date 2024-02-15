package keystore

import (
	"crypto/rand"
	"ctb-cli/crypto/key_crypto"
	"ctb-cli/types"
	"io"
)

func (ks *KeyStoreDefault) LoadKeys() error {

	if ks.privateKey != nil {
		return nil
	}
	serializedPrivateKey, err := ks.keyRepository.GetPrivateKey()
	if err != nil {
		return err
	}
	ks.privateKey, err = key_crypto.OpenPrivateKey(serializedPrivateKey, &ks.rootKey)
	if err != nil {
		return err
	}
	return nil
}

func (ks *KeyStoreDefault) GenerateClientKeys() (err error) {
	//Generate private key
	privateKey := types.Key{}
	io.ReadFull(rand.Reader, privateKey[:])
	//Save private key
	sealPrivateKey, err := key_crypto.SealPrivateKey(privateKey[:], &ks.rootKey)
	if err != nil {
		return err
	}
	err = ks.keyRepository.SavePrivateKey(sealPrivateKey)
	if err != nil {
		return err
	}

	publicKey, err := ks.getPublicKey()
	if err != nil {
		return err
	}
	serializedPublic, err := key_crypto.SerializePublicKey(publicKey)
	if err != nil {
		return err
	}
	err = ks.keyRepository.SavePublicKey(ks.clintId, serializedPublic)
	return
}
