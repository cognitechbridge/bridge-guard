package file_crypto

import (
	"ctb-cli/core"
	"ctb-cli/crypto/stream"
	"io"
)

// writer is a struct for generating encrypted files.
type writer struct {
	header       Header
	notFirst     bool
	dst          io.Writer
	streamWriter *stream.Writer
}

var (
	fileVersion = []byte{1} // Define the file version
)

// NewWriter creates a new writer.
func NewWriter(dst io.Writer, keyInfo *core.KeyInfo, fileId string) (*writer, error) {
	streamWriter, err := stream.NewWriter(keyInfo.Key[:], dst)
	if err != nil {
		return nil, err
	}
	return &writer{
		dst:          dst,
		header:       newHeader(fileId, keyInfo.Id),
		notFirst:     false,
		streamWriter: streamWriter,
	}, nil
}

func (e *writer) Write(buf []byte) (int, error) {
	if !e.notFirst {
		if err := e.writeFileVersionAndHeader(); err != nil {
			return 0, err
		}
		e.notFirst = true
	}
	return e.streamWriter.Write(buf)
}

// writeFileVersionAndHeader appends the header to the buffer.
func (e *writer) writeFileVersionAndHeader() (err error) {
	_, err = e.dst.Write(fileVersion)
	if err != nil {
		return err
	}
	headerBytes, err := e.header.Marshal()
	if err != nil {
		return err
	}
	_, err = e.dst.Write(headerBytes)
	return err
}

func (e *writer) Close() error {
	if !e.notFirst {
		if err := e.writeFileVersionAndHeader(); err != nil {
			return err
		}
		e.notFirst = true
	}
	return e.streamWriter.Close()
}

func newHeader(fileId string, keyId string) Header {
	return Header{
		Version: "V1",
		Alg:     getAlgorithmName(), // Set default algorithm
		FileID:  fileId,
		KeyId:   keyId,
	}
}
