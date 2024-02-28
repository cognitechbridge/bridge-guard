package file_crypto

import (
	"ctb-cli/crypto/stream"
	"ctb-cli/core"
	"errors"
	"io"
)

type EncryptedStream struct {
	source io.Reader
}

func Parse(source io.Reader) (*Header, *EncryptedStream, error) {
	header, err := readFileVersionAndHeader(source)
	if err != nil {
		return nil, nil, err
	}
	return header, &EncryptedStream{source: source}, nil
}

func (e EncryptedStream) Decrypt(key *core.Key) (io.Reader, error) {
	return stream.NewReader(key[:], e.source)
}

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
