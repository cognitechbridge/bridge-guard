package keystore

import (
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"os"
)

func (ks *KeyStore) ReadRecoveryKey(inPath string) error {
	path := os.ExpandEnv(inPath)
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil || block.Type != "PUBLIC KEY" {
		return err
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	pubKey, ok := pub.(*rsa.PublicKey)

	if !ok {
		return errors.New("public key is not of type RSA Public Key")
	}

	ks.recoveryPublicKey = pubKey

	pubASN1, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return err
	}
	hash := sha1.Sum(pubASN1)
	ks.recoverySha1 = hex.EncodeToString(hash[:])

	return nil
}
