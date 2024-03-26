package file_crypto_test

import (
	"bytes"
	"crypto/rand"
	"ctb-cli/core"
	"ctb-cli/crypto/file_crypto"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	// Generate some random data
	originalData := make([]byte, 1024)
	_, err := rand.Read(originalData)
	if err != nil {
		t.Fatal(err)
	}

	// Create a key
	key := core.KeyInfo{
		Id:  "ID",
		Key: make([]byte, 32),
	}

	// Create an in-memory read-write buffer
	memBuf := bytes.NewBuffer(nil)

	// Create a writer that writes to the in-memory buffer
	memEncryptedWriter, err := file_crypto.NewWriter(memBuf, &key, "fileId")
	if err != nil {
		t.Fatal(err)
	}

	// Write the data to the encrypted writer
	n, err := memEncryptedWriter.Write(originalData)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(originalData) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(originalData), n)
	}

	// Close the writer to finalize the encryption
	err = memEncryptedWriter.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Create a parser to read the data back
	header, encStream, err := file_crypto.Parse(memBuf)
	if err != nil {
		t.Fatal(err)
	}

	// Assert header values
	if header.FileID != "fileId" {
		t.Errorf("Expected FileID to be 'fileId', got '%s'", header.FileID)
	}
	if header.KeyId != "ID" {
		t.Errorf("Expected KeyID to be 'ID', got '%s'", header.KeyId)
	}

	// Read the data back
	readData := make([]byte, 1024)
	decryptedData, err := encStream.Decrypt(&key)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = decryptedData.Read(readData)

	// Check if the original and read data are the same
	if !bytes.Equal(originalData, readData) {
		t.Errorf("Original and read data do not match")
	}
}
