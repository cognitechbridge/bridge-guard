package keystore

import (
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
)

func (ks *KeyStore) ReadRecoveryKey(inPath string) error {
	path := os.ExpandEnv(inPath)

	// Load the PEM file
	pemData, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Failed to read file: %s\n", err)
	}

	// Decode the PEM file
	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "CERTIFICATE" {
		return fmt.Errorf("Failed to decode PEM block containing certificate\n")
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("Failed to parse certificate: %s\n", err)
	}

	rsaPubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("Public key is not of RSA type\n")
	}

	ks.recoveryPublicKey = rsaPubKey

	sha1Hash := sha1.Sum(cert.Raw)
	ks.recoverySha1 = hex.EncodeToString(sha1Hash[:])

	return nil
}
