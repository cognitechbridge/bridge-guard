package file_crypto

import (
	"ctb-cli/crypto/stream"
	"ctb-cli/types"
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
func NewWriter(dst io.Writer, keyInfo *types.KeyInfo, clientId string, fileId string) (*writer, error) {
	streamWriter, err := stream.NewWriter(keyInfo.Key[:], dst)
	if err != nil {
		return nil, err
	}
	return &writer{
		dst:          dst,
		header:       newHeader(clientId, fileId, keyInfo.Id, keyInfo.RecoveryBlobs),
		notFirst:     false,
		streamWriter: streamWriter,
	}, nil
}

func (e *writer) Write(buf []byte) (int, error) {
	if e.notFirst == false {
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
	_, err = e.dst.Write(headerBytes)
	return err
}

func (e *writer) Close() error {
	return e.streamWriter.Close()
}

func newHeader(clientId string, fileId string, keyId string, recoveryBlobs []string) Header {
	return Header{
		Version:    "V1",
		Alg:        getAlgorithmName(), // Set default algorithm
		ClientID:   clientId,
		FileID:     fileId,
		KeyId:      keyId,
		Recoveries: recoveryBlobs,
	}
}
