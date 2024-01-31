/*
Copyright © 2024 Mohammad Saadatfar
*/
package main

import "ctb-cli/cmd"

func main() {

	//var key encryptor.Key

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
	cmd.Execute()

	//
	//if err != nil {
	//	fmt.Println("Error reading managerConfig", err)
	//}
	//
	//chunkSize := viper.GetUint64("crypto.chunk-size")
	//viper.Set("crypto.chunk-size", 100)
	//if err != nil {
	//	fmt.Println("Error reading managerConfig", err)
	//}
	//fmt.Println("Error reading managerConfig", chunkSize)

	//Replace with your actual encryption key and nonce
	//var key encryptor.Key
	//
	//s3Client := s3.NewClient("ctb-test-2", 10*1024*1024)

	//cloudClient := cloud.NewClient("http://localhost:1323", 10*1024*1024)
	//cloudClient := file_db.NewDummyClient()

	//sqlLiteConnection, _ := db.NewSqlLiteConnection()
	//
	//keyStore := keystore.NewKeyStore(key, sqlLiteConnection)
	//path, err := config.Crypto.GetRecoveryPublicCertPath()
	//if err != nil {
	//	return
	//}
	//err = keyStore.ReadRecoveryKey(path)
	//if err != nil {
	//	fmt.Println("Error reading crt:", err)
	//	return
	//}
	//
	//filesystem := filesyetem.NewPersistFileSystem(sqlLiteConnection)
	//
	//seed := time.Now().UTC().UnixNano()
	//nameGenerator := namegenerator.NewNameGenerator(seed)
	//name := nameGenerator.Generate()
	//
	//chunkSize, _ := config.Crypto.GetChunkSize()
	//clientId, _ := config.Workspace.GetClientId()
	//managerConfig := manager.Config{
	//	EncryptChunkSize: chunkSize,
	//	ClientId:         clientId,
	//}
	//managerClient := manager.NewManager(
	//	managerConfig,
	//	keyStore,
	//	filesystem,
	//	cloudClient,
	//)
	//
	//manager.Client.Init(
	//	managerConfig,
	//	keyStore,
	//	filesystem,
	//	cloudClient,
	//)

	//fmt.Println("upload started")
	//startTime := time.Now()
	//
	//uploader := managerClient.NewUploader("D:\\sample.txt", name)
	//_, err = uploader.upload()
	//if err != nil {
	//	fmt.Println("Encryption failed:", err)
	//}
	//elapsedTime := time.Since(startTime)
	//fmt.Printf("upload took %s\n", elapsedTime)

	//downloader := manager_client.NewDownloader("D:\\unencrypted.txt", name)
	//err = downloader.Download()
	//if err != nil {
	//	fmt.Println("Encryption failed:", err)
	//}
	//
	//fmt.Println("Decryption complete and data written to file.")
}
