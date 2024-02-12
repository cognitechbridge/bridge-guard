package file_crypto

import (
	"ctb-cli/crypto/stream"
	"ctb-cli/types"
	"encoding/json"
	"io"
)

// writer is a struct for generating encrypted files.
type writer struct {
	header       fileHeader
	notFirst     bool
	dst          io.Writer
	streamWriter *stream.Writer
}

// NewWriter creates a new writer.
func NewWriter(dst io.Writer, key types.Key, clientId string, fileId string, recoveryBlobs []string) (*writer, error) {
	streamWriter, err := stream.NewWriter(key[:], dst)
	if err != nil {
		return nil, err
	}
	return &writer{
		dst:          dst,
		header:       newEncryptedFileHeader(clientId, fileId, recoveryBlobs),
		notFirst:     false,
		streamWriter: streamWriter,
	}, nil
}

func (e *writer) Write(buf []byte) (int, error) {
	if e.notFirst == false {
		if err := e.appendHeader(); err != nil {
			return 0, err
		}
		e.notFirst = true
	}
	return e.streamWriter.Write(buf)
}

// appendHeader appends the header to the buffer.
func (e *writer) appendHeader() (err error) {
	_, err = e.dst.Write(fileVersion)
	if err != nil {
		return err
	}
	headerBytes, err := json.Marshal(e.header)
	if err != nil {
		return err
	}
	err = e.writeContext(string(headerBytes))
	return err
}

// writeContext writes a string to the buffer with its length.
func (e *writer) writeContext(context string) (err error) {
	contextLength := len(context)
	// Assumes context length fits in 2 bytes
	_, err = e.dst.Write([]byte{byte(contextLength >> 8), byte(contextLength)})
	if err != nil {
		return
	}
	_, err = e.dst.Write([]byte(context))
	if err != nil {
		return
	}
	return nil
}

func (e *writer) Close() error {
	return e.streamWriter.Close()
}
