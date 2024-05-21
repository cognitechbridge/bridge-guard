package file_crypto

import (
	"ctb-cli/core"
	"ctb-cli/crypto/stream"
	"errors"
	"io"
)

// writer represents a writer that performs cryptographic operations on a file.
type writer struct {
	header       Header         // The header of the file.
	notFirst     bool           // Indicates whether it is not the first write operation.
	dst          io.Writer      // The destination writer to write the encrypted data to.
	streamWriter *stream.Writer // The stream writer used for encryption.
}

var (
	fileVersion = []byte{1} //Current encryption file version
)

// NewWriter creates a new writer object that encrypts data and writes it to the specified destination writer.
// It takes the destination writer, key information, and file ID as parameters.
// The function returns a pointer to the writer object and an error if any occurred during the creation process.
func NewWriter(dst io.Writer, keyInfo *core.KeyInfo, fileId string) (*writer, error) {
	// Create a new stream writer with the key and the destination writer.
	streamWriter, err := stream.NewWriter(keyInfo.Key.Bytes(), dst)
	if err != nil {
		return nil, err
	}
	// Create a new writer object with the destination writer, header, and stream writer.
	return &writer{
		dst:          dst,
		header:       newHeader(fileId, keyInfo.Id),
		notFirst:     false,
		streamWriter: streamWriter,
	}, nil
}

// Write writes the given byte slice to the underlying stream.
// It first checks if it's the first write operation, and if so, it writes the file version and header.
// Returns the number of bytes written and any error encountered.
func (e *writer) Write(buf []byte) (int, error) {
	if !e.notFirst {
		// Write the file version and header if it's the first write operation.
		if err := e.writeFileVersionAndHeader(); err != nil {
			return 0, err
		}
		e.notFirst = true
	}
	return e.streamWriter.Write(buf)
}

// writeFileVersionAndHeader writes the file version and header to the destination writer.
// It returns an error if there was a problem writing the version or header.
func (e *writer) writeFileVersionAndHeader() (err error) {
	// write file version.
	_, err = e.dst.Write(fileVersion)
	if err != nil {
		return err
	}
	// Marshal the header and write it to the destination writer.
	headerBytes, err := e.header.Marshal()
	if err != nil {
		return err
	}
	_, err = e.dst.Write(headerBytes)
	return err
}

// Close closes the writer and finalizes the encryption.
// If it's the first write operation, it writes the file version and header before closing.
// This is to handle the case where the writer is closed without any write operation (e.g. empty file).
// It returns an error if there was an error writing the file version and header, or if there was an error closing the stream writer.
func (e *writer) Close() error {
	if !e.notFirst {
		// Write the file version and header if it's the first write operation.
		// This is to handle the case where the writer is closed without any write operation. (e.g. empty file)
		if err := e.writeFileVersionAndHeader(); err != nil {
			return err
		}
		e.notFirst = true
	}
	// Close the stream writer to flush the remaining data and finalize the encryption.
	return e.streamWriter.Close()
}

// newHeader creates a new Header struct with the specified fileId and keyId.
// It sets the default algorithm by calling the getAlgorithmName function.
func newHeader(fileId string, keyId string) Header {
	return Header{
		Version: "V1",
		Alg:     "AEAD_ChaCha20_Poly1305", // Set default algorithm
		FileID:  fileId,
		KeyId:   keyId,
	}
}

// EncryptedStream represents an encrypted stream of data.
type EncryptedStream struct {
	source io.Reader
}

// Parse reads the encrypted data from the provided source and returns the parsed header,
// an encrypted stream, and any error encountered during the process.
func Parse(source io.Reader) (*Header, *EncryptedStream, error) {
	header, err := readFileVersionAndHeader(source)
	if err != nil {
		return nil, nil, err
	}
	return header, &EncryptedStream{source: source}, nil
}

// Decrypt decrypts the encrypted stream using the provided key.
// It returns an io.Reader that can be used to read the decrypted data.
// If an error occurs during decryption, it is returned along with nil reader.
func (e EncryptedStream) Decrypt(key *core.KeyInfo) (io.Reader, error) {
	return stream.NewReader(key.Key.Bytes(), e.source)
}

// readFileVersionAndHeader reads the file version and header from the given source.
// It returns the parsed header and any error encountered during the process.
func readFileVersionAndHeader(source io.Reader) (*Header, error) {
	err := readFileVersion(source)
	if err != nil {
		return nil, err
	}
	header, err := ParseHeader(source)
	if err != nil {
		return nil, err
	}
	return header, err
}

// readFileVersion reads the version byte from the given source.
// It returns an error if the version is not supported.
// The version byte is expected to be the first byte in the source.
// The current version is 1.
func readFileVersion(source io.Reader) error {
	// Create a buffer to hold the version byte
	versionBuffer := make([]byte, 1)
	_, err := io.ReadFull(source, versionBuffer)
	if err != nil {
		return err
	}
	version := versionBuffer[0]

	// Check the version (assuming version 1 is expected)
	if version != 1 {
		return errors.New("unsupported file version")
	}
	return err
}
