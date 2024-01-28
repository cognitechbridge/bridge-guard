/*
Copyright © 2024 Mohammad Saadatfar
*/
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"ctb-cli/db"
	"ctb-cli/encryptor"
	"ctb-cli/file_db/cloud"
	"ctb-cli/filesyetem"
	"ctb-cli/keystore"
	"ctb-cli/secure_storage"
	"encoding/pem"
	"fmt"
	"github.com/goombaio/namegenerator"
	"github.com/spf13/viper"
	"math/big"
	"os"
	"time"
)

func savePublicKeyAsCertificate(fileName string, pubkey *rsa.PublicKey, privKey *rsa.PrivateKey) error {
	// Set up a template for the certificate
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	certTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"CTB"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year validity
		KeyUsage:              x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{},
		BasicConstraintsValid: true,
	}

	// Create the certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, pubkey, privKey)
	if err != nil {
		return err
	}

	// Encode the certificate to PEM format
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})

	// Write the certificate to the file
	return os.WriteFile(fileName, certPEM, 0600)
}

func main() {

	var key encryptor.Key

	// Generate RSA keys
	//const keySize = 2048
	//privKey, err := rsa.GenerateKey(rand.Reader, keySize)
	//if err != nil {
	//	fmt.Fprintf(os.Stderr, "Error generating RSA key: %s\n", err)
	//	return
	//}
	//pubKey := &privKey.PublicKey

	//savePublicKeyAsCertificate("D:/a.crt", pubKey, privKey)

	//fmt.Println("Decrypted Message: ", decryptedMessage)

	//cmd.Execute()
	//
	viper.SetConfigName("config")            // name of config file (without extension)
	viper.SetConfigType("yaml")              // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/.ctb/")        // path to look for the config file in
	viper.AddConfigPath("$USERPROFILE/.ctb") // call multiple times to add many search paths
	viper.AddConfigPath(".")                 // optionally look for config in the working directory
	err := viper.ReadInConfig()
	//
	//if err != nil {
	//	fmt.Println("Error reading config", err)
	//}
	//
	//chunkSize := viper.GetUint64("crypto.chunk-size")
	//viper.Set("crypto.chunk-size", 100)
	//if err != nil {
	//	fmt.Println("Error reading config", err)
	//}
	//fmt.Println("Error reading config", chunkSize)

	//Replace with your actual encryption key and nonce
	//var key encryptor.Key
	//
	//s3Client := s3.NewClient("ctb-test-2", 10*1024*1024)
	cloudClient := cloud.NewClient("http://localhost:1323", 10*1024*1024)
	//cloudClient := file_db.NewDummyClient()

	sqlLiteConnection, _ := db.NewSqlLiteConnection()
	keyStore := keystore.NewKeyStore(key, sqlLiteConnection)
	err = keyStore.ReadRecoveryKey(
		viper.GetString("crypto.recovery-public-cert"),
	)
	if err != nil {
		fmt.Println("Error reading crt:", err)
		return
	}
	filesystem := filesyetem.NewPersistFileSystem(sqlLiteConnection)

	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)
	name := nameGenerator.Generate()

	config := secure_storage.ManagerConfig{
		EncryptChunkSize: 1024 * 1024,
	}
	manager := secure_storage.NewManager(
		config,
		keyStore,
		filesystem,
		cloudClient,
	)

	fmt.Println("Upload started")
	startTime := time.Now()
	uploader := manager.NewUploader("D:\\sample.txt", name)
	_, err = uploader.Upload()
	if err != nil {
		fmt.Println("Encryption failed:", err)
	}
	elapsedTime := time.Since(startTime)
	fmt.Printf("Upload took %s\n", elapsedTime)

	//downloader := manager.NewDownloader("D:\\unencrypted.txt", name)
	//err = downloader.Download()
	//if err != nil {
	//	fmt.Println("Encryption failed:", err)
	//}
	//
	//fmt.Println("Decryption complete and data written to file.")
}
