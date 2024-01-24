package encryptor

import (
	"encoding/json"
	"io"
)

var (
	Separator            = []byte{0, 0, 0, 0} // Define the appropriate separator
	EncryptedFileVersion = []byte{1}          // Define the file version
)

// EncryptedFileGenerator is a struct for generating encrypted files.
type EncryptedFileGenerator struct {
	source       io.Reader
	header       EncryptionFileHeader
	key          Key
	buffer       []byte
	nonce        Nonce
	chunkCounter uint32
	chunkSize    uint64
}

// NewEncryptedFileGenerator creates a new EncryptedFileGenerator.
func NewEncryptedFileGenerator(source io.Reader, key Key, chunkSize uint64, clientId string, fileId string, recoveryBlob string) *EncryptedFileGenerator {
	return &EncryptedFileGenerator{
		source:       source,
		header:       NewEncryptionFileHeader(chunkSize, clientId, fileId, recoveryBlob),
		buffer:       make([]byte, 0),
		chunkCounter: 0,
		key:          key,
		nonce:        Nonce{},
		chunkSize:    chunkSize,
	}
}

// Read implements the io.Reader interface for EncryptedFileGenerator.
func (efg *EncryptedFileGenerator) Read(buf []byte) (int, error) {
	if efg.chunkCounter == 0 {
		if err := efg.appendHeader(); err != nil {
			return 0, err
		}
		efg.chunkCounter++
	}

	for len(efg.buffer) < len(buf) {
		encryptedBytes, err := efg.readBytesEncrypted()
		if err != nil && err != io.EOF {
			return 0, err
		}
		if err == io.EOF || len(encryptedBytes) == 0 {
			break
		}
		efg.buffer = append(efg.buffer, Separator...)
		efg.buffer = append(efg.buffer, encryptedBytes...)
		efg.chunkCounter++
		efg.nonce.increaseBe()
	}

	if len(efg.buffer) == 0 {
		return 0, io.EOF
	}

	n := copy(buf, efg.buffer)
	efg.buffer = efg.buffer[n:]

	return n, nil // Data read successfully
}

// appendHeader appends the header to the buffer.
func (efg *EncryptedFileGenerator) appendHeader() error {
	efg.buffer = append(efg.buffer, EncryptedFileVersion...)
	headerBytes, err := json.Marshal(efg.header)
	if err != nil {
		return err
	}
	efg.writeContext(string(headerBytes))
	return nil
}

// writeContext writes a string to the buffer with its length.
func (efg *EncryptedFileGenerator) writeContext(context string) {
	contextLength := len(context)
	// Assumes context length fits in 2 bytes
	efg.buffer = append(efg.buffer, byte(contextLength), byte(contextLength>>8))
	efg.buffer = append(efg.buffer, context...)
}

// readBytesEncrypted reads and encrypts a chunk of data from the source.
func (efg *EncryptedFileGenerator) readBytesEncrypted() ([]byte, error) {
	buffer := make([]byte, efg.chunkSize)
	n, err := efg.source.Read(buffer)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}
	return EncryptChunk(buffer[:n], efg.key, efg.nonce)
}
