package recovery

import (
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"ctb-cli/core"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
)

func UnmarshalRecoveryItem(pemBytes []byte) (*core.RecoveryItem, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("invalid public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pubKey, ok := pub.(*rsa.PublicKey)

	if !ok {
		return nil, errors.New("public key is not of type RSA Public Key")
	}

	rec := core.RecoveryItem{}

	rec.PublicKey = pubKey

	pubASN1, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, err
	}
	hash := sha1.Sum(pubASN1)
	rec.Sha1 = hex.EncodeToString(hash[:])

	return &rec, nil
}
