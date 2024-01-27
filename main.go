package main

import (
	"fmt"
	"github.com/goombaio/namegenerator"
	"storage-go/encryptor"
	"storage-go/filesyetem"
	"storage-go/keystore"
	"storage-go/persist"
	"storage-go/persist_file"
	"storage-go/secure_storage"
	"time"
)

func main() {
	// Replace with your actual encryption key and nonce
	var key encryptor.Key

	//s3storage := storage.NewS3Client("ctb-test-2", 10*1024*1024)
	cloudClient := persist_file.NewCtbCloudClient("http://localhost:1323", 10*1024*1024)
	//cloudClient := persist_file.NewDummyClient()

	sqlLiteConnection, _ := persist.NewSqlLiteConnection()
	keyStore := keystore.NewKeyStore(key, sqlLiteConnection)
	filesystem := filesyetem.NewPersistFileSystem(sqlLiteConnection)

	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)
	name := nameGenerator.Generate()

	manager := secure_storage.NewManager(keyStore, filesystem, cloudClient)

	fmt.Println("Upload started")
	startTime := time.Now()
	uploader := manager.NewUploader("D:\\sample.txt", name)
	_, err := uploader.Upload()
	if err != nil {
		fmt.Println("Encryption failed:", err)
	}
	elapsedTime := time.Since(startTime)
	fmt.Printf("Upload took %s\n", elapsedTime)

	downloader := manager.NewDownloader("D:\\unencrypted.txt", name)
	err = downloader.Download()
	if err != nil {
		fmt.Println("Encryption failed:", err)
	}

	fmt.Println("Decryption complete and data written to file.")
}
