/*
Copyright © 2024 Mohammad Saadatfar
*/
package main

import (
	"ctb-cli/cmd"
)

func main() {

	cmd.Execute()

	//Replace with your actual encryption key and nonce
	//var key encryptor.Key
	//
	////s3Client := s3.NewClient("ctb-test-2", 10*1024*1024)
	//cloudClient := cloud.NewClient("http://localhost:1323", 10*1024*1024)
	////cloudClient := file_db.NewDummyClient()
	//
	//sqlLiteConnection, _ := db.NewSqlLiteConnection()
	//keyStore := keystore.NewKeyStore(key, sqlLiteConnection)
	//filesystem := filesyetem.NewPersistFileSystem(sqlLiteConnection)
	//
	//seed := time.Now().UTC().UnixNano()
	//nameGenerator := namegenerator.NewNameGenerator(seed)
	//name := nameGenerator.Generate()
	//
	//config := secure_storage.ManagerConfig{
	//	EncryptChunkSize: 1024 * 1024,
	//}
	//manager := secure_storage.NewManager(
	//	config,
	//	keyStore,
	//	filesystem,
	//	cloudClient,
	//)
	//
	//fmt.Println("Upload started")
	//startTime := time.Now()
	//uploader := manager.NewUploader("D:\\sample.txt", name)
	//_, err := uploader.Upload()
	//if err != nil {
	//	fmt.Println("Encryption failed:", err)
	//}
	//elapsedTime := time.Since(startTime)
	//fmt.Printf("Upload took %s\n", elapsedTime)
	//
	//downloader := manager.NewDownloader("D:\\unencrypted.txt", name)
	//err = downloader.Download()
	//if err != nil {
	//	fmt.Println("Encryption failed:", err)
	//}
	//
	//fmt.Println("Decryption complete and data written to file.")
}
