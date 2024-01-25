package main

import (
	"fmt"
	"github.com/goombaio/namegenerator"
	"storage-go/encryptor"
	"storage-go/filesyetem"
	"storage-go/keystore"
	"storage-go/persist"
	"storage-go/secure_storage"
	"storage-go/storage"
	"time"
)

func main() {
	// Replace with your actual encryption key and nonce
	var key encryptor.Key

	s3storage := storage.NewS3Storage("ctb-test-2", 10*1024*1024)
	sqlLiteConnection, _ := persist.NewSqlLiteConnection()
	keyStore := keystore.NewKeyStore(key, sqlLiteConnection)
	filesystem := filesyetem.NewPersistFileSystem(sqlLiteConnection)

	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)
	name := nameGenerator.Generate()

	manager := secure_storage.NewManager(keyStore, s3storage, filesystem)

	uploader := manager.NewUploader("D:\\sample.txt", name)
	_, err := uploader.Download()
	if err != nil {
		fmt.Println("Encryption failed:", err)
	}

	downloader := manager.NewDownloader("D:\\unencrypted.txt", name)
	err = downloader.Download()
	if err != nil {
		fmt.Println("Encryption failed:", err)
	}

	fmt.Println("Decryption complete and data written to file.")
}
