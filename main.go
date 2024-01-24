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
	x := keystore.NewKeyStore(key, sqlcon)
	x.Insert("Test", key)

	err := encrypt(key)
	if err != nil {
		fmt.Println("Encryption failed:", err)
	}
	err = decrypt(key)
	if err != nil {
		fmt.Println("Encryption failed:", err)
	}

	fmt.Println("Decryption complete and data written to file.")
}

func encrypt(key encryptor.Key) error {
	inputFile, err := os.Open("D:\\sample.txt")
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create("D:\\encrypted.txt")
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	fileUuid, _ := uuid.NewV7()
	clientId := "CLIENTID"

	efg := encryptor.NewEncryptedFileGenerator(inputFile, key, 10*1024*1024, clientId, fileUuid.String(), "")

	_, err = io.Copy(outputFile, efg)
	if err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	return nil
}

func decrypt(key encryptor.Key) error {
	// Open the encrypted file
	file, err := os.Open("D:\\encrypted.txt")
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Create a new ReaderDecryptor
	rd, err := encryptor.NewReaderDecryptor(key, file)
	if err != nil {
		return fmt.Errorf("error creating ReaderDecryptor: %w", err)
	}

	// Create or open the output file
	outputFile, err := os.Create("D:\\decrypted.txt") // Specify your output file path
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, rd)
	if err != nil {
		return fmt.Errorf("error Copying file: %w", err)
	}

	return nil
}
