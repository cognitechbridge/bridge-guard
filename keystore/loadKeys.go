package keystore

import (
	"crypto/rand"
	"ctb-cli/types"
	"fmt"
	"io"
)

func (ks *KeyStore) LoadKeys() error {
	//if err := ks.loadPublicKey(); err != nil {
	//	return err
	//}

	if ks.privateKey != nil {
		return nil
	}
	serializedPrivateKey, err := ks.persist.GetPrivateKey()
	if err != nil {
		return err
	}
	privateKey, err := ks.DeserializePrivateKey(serializedPrivateKey, &ks.rootKey)
	fmt.Printf("%v", privateKey)
	//if err != nil {
	//	return err
	//}
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
	privateKey := types.Key{}
	io.ReadFull(rand.Reader, privateKey[:])
	//Save private key
	serialized, err := ks.SerializePrivateKey(privateKey[:], &ks.rootKey)
	if err != nil {
		return err
	}
	err = ks.persist.SavePrivateKey(serialized)
	if err != nil {
		return err
	}
	return nil
}
