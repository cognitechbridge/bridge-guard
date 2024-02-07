package keystore

import (
	"crypto/rand"
	"crypto/rsa"
)

func (ks *KeyStore) LoadKeys() error {
	if err := ks.loadPublicKey(); err != nil {
		return err
	}

	if ks.privateKey != nil {
		return nil
	}
	privateKey, err := ks.persist.GetPrivateKey()
	if err != nil {
		return err
	}
	ks.privateKey, err = ks.DeserializePrivateKey(privateKey, &ks.rootKey)
	if err != nil {
		return err
	}
	return nil
}

func (ks *KeyStore) loadPublicKey() error {
	pub, err := ks.persist.GetPublicKey(ks.clintId)
	if err != nil {
		return err
	}
	ks.publicKey = pub
	return nil
}

func (ks *KeyStore) GenerateClientKeys() (err error) {
	//Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	//Generate public key
	publicKey := &privateKey.PublicKey
	//Save public key
	err = ks.persist.SavePublicKey(ks.clintId, publicKey)
	if err != nil {
		return err
	}
	//Save private key
	serialized, err := ks.SerializePrivateKey(privateKey, &ks.rootKey)
	if err != nil {
		return err
	}
	err = ks.persist.SavePrivateKey(serialized)
	if err != nil {
		return err
	}
	return nil
}
