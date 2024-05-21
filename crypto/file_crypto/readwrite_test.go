package file_crypto_test

import (
	"bytes"
	"crypto/rand"
	"ctb-cli/core"
	"ctb-cli/crypto/file_crypto"
	"io"
	"testing"
)

func testRoundTrip(t *testing.T, length int) {
	// Generate some random data
	originalData := make([]byte, length)
	if length != 0 {
		_, _ = rand.Read(originalData)
	}

	// Create a key
	keyInfo := core.KeyInfo{
		Id:  "ID",
		Key: core.NewKeyFromRand(),
	}

	// Create an in-memory read-write buffer
	memBuf := bytes.NewBuffer(nil)

	// Create a writer that writes to the in-memory buffer
	memEncryptedWriter, err := file_crypto.NewWriter(memBuf, &keyInfo, "fileId")
	if err != nil {
		t.Fatal(err)
	}

	// Write the data to the encrypted writer
	if length > 0 {
		n, err := memEncryptedWriter.Write(originalData)
		if err != nil {
			t.Fatal(err)
		}
		if n != len(originalData) {
			t.Errorf("Expected to write %d bytes, wrote %d", len(originalData), n)
		}
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
	if header.Alg != "AEAD_ChaCha20_Poly1305" {
		t.Errorf("Expected Alg to be 'AEAD_ChaCha20_Poly1305', got '%s'", header.Alg)
	}
	if header.Version != "V1" {
		t.Errorf("Expected Version to be V1, got %s", header.Version)
	}

	// Read the data back
	decryptedData, err := encStream.Decrypt(&keyInfo)
	if err != nil {
		t.Fatal(err)
	}
	readData, err := io.ReadAll(decryptedData)
	if err != nil {
		t.Fatal(err)
	}
	if len(readData) != length {
		t.Errorf("Expected to read %d bytes, read %d", length, len(readData))
	}

	// Check if the original and read data are the same
	if !bytes.Equal(originalData, readData) {
		t.Errorf("Original and read data do not match")
	}
}

// TestRoundTrip tests the round trip of writing and reading encrypted data
func TestRoundTrip(t *testing.T) {
	testRoundTrip(t, 0)
	testRoundTrip(t, 1024)
	testRoundTrip(t, 1024*1024)
}
