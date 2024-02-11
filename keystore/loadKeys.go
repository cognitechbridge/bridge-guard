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
	sealPrivateKey, err := ks.SealPrivateKey(privateKey[:], &ks.rootKey)
	if err != nil {
		return err
	}
	err = ks.persist.SavePrivateKey(sealPrivateKey)
	if err != nil {
		return err
	}

	publicKey, err := ks.getPublicKey()
	if err != nil {
		return err
	}
	serializedPublic, err := ks.SerializePublicKey(publicKey)
	if err != nil {
		return err
	}
	err = ks.persist.SavePublicKey(ks.clintId, serializedPublic)
	return
}
