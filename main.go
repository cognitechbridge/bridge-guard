package main

import (
	"fmt"
	"github.com/google/uuid"
	"io"
	"os"
	"storage-go/encryptor"
	"storage-go/keystore"
	"storage-go/persist"
)

func main() {
	// Replace with your actual encryption key and nonce
	var key encryptor.Key

	sqlcon, _ := persist.NewSqlLiteConnection()
	keyStore := keystore.NewKeyStore(key, sqlcon)

	fileId, err := encrypt(keyStore)
	if err != nil {
		fmt.Println("Encryption failed:", err)
	}
	err = decrypt(fileId, keyStore)
	if err != nil {
		fmt.Println("Encryption failed:", err)
	}

	fmt.Println("Decryption complete and data written to file.")
}

func encrypt(store *keystore.KeyStore) (string, error) {
	//Open input file
	inputFile, err := os.Open("D:\\sample.txt")
	if err != nil {
		return "", fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	//Create output file
	outputFile, err := os.Create("D:\\encrypted.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	//Create header parameters
	clientId := "CLIENTID"
	fileUuid, _ := uuid.NewV7()
	pair, err := store.GenerateKeyPair(fileUuid.String())
	if err != nil {
		return "", err
	}

	//Create reader
	efg := encryptor.NewEncryptedFileGenerator(
		inputFile,
		pair.Key,
		10*1024*1024,
		clientId,
		fileUuid.String(),
		pair.RecoveryBlob,
	)

	//Copy to output
	_, err = io.Copy(outputFile, efg)
	if err != nil {
		return "", fmt.Errorf("error copying file: %w", err)
	}

	//Return file id
	return fileUuid.String(), nil
}

func decrypt(id string, store *keystore.KeyStore) error {
	// Open the encrypted file
	file, err := os.Open("D:\\encrypted.txt")
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	//Get data key
	dataKey, err := store.Get(id)
	if err != nil {
		return fmt.Errorf("error creating ReaderDecryptor: %w", err)
	}

	// Create a new ReaderDecryptor
	rd, err := encryptor.NewReaderDecryptor(*dataKey, file)
	if err != nil {
		return fmt.Errorf("error creating ReaderDecryptor: %w", err)
	}

	// Create or open the output file
	outputFile, err := os.Create("D:\\decrypted.txt") // Specify your output file path
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer outputFile.Close()

	//Copy to output
	_, err = io.Copy(outputFile, rd)
	if err != nil {
		return fmt.Errorf("error Copying file: %w", err)
	}

	return nil
}
