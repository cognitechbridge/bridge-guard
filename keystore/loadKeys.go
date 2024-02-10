package keystore

import (
	"crypto/rand"
	"ctb-cli/types"
	"io"
)

func (ks *KeyStore) LoadKeys() error {

	if ks.privateKey != nil {
		return nil
	}
	serializedPrivateKey, err := ks.persist.GetPrivateKey()
	if err != nil {
		return err
	}
	ks.privateKey, err = ks.OpenPrivateKey(serializedPrivateKey, &ks.rootKey)
	if err != nil {
		return err
	}
	return nil
}

func (ks *KeyStore) GenerateClientKeys() (err error) {
	//Generate private key
	privateKey := types.Key{}
	io.ReadFull(rand.Reader, privateKey[:])
	//Save private key
	serialized, err := ks.SealPrivateKey(privateKey[:], &ks.rootKey)
	if err != nil {
		return err
	}
	err = ks.persist.SavePrivateKey(serialized)
	if err != nil {
		return err
	}
	return nil
}
