package file_crypto

import (
	"ctb-cli/crypto/stream"
	"ctb-cli/types"
	"errors"
	"io"
)

type Reader struct {
	streamReader *stream.Reader
}

func Parse(key *types.Key, source io.Reader) (*Header, *Reader, error) {
	streamReader, err := stream.NewReader(key[:], source)
	if err != nil {
		return nil, nil, err
	}
	header, err := readFileVersionAndHeader(source)
	return header, &Reader{streamReader: streamReader}, nil
}

func (d *Reader) Read(p []byte) (int, error) {
	return d.streamReader.Read(p)
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
