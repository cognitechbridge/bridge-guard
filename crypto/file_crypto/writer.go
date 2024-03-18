package file_crypto

import (
	"ctb-cli/core"
	"ctb-cli/crypto/stream"
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
	streamWriter, err := stream.NewWriter(keyInfo.Key[:], dst)
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
		Alg:     getAlgorithmName(), // Set default algorithm
		FileID:  fileId,
		KeyId:   keyId,
	}
}
