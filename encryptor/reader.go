package encryptor

import (
	"ctb-cli/encryptor/stream"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
)

type Reader struct {
	source       io.Reader
	header       *EncryptedFileHeader
	notFirst     bool
	streamReader *stream.Reader
}

func NewReader(key *Key, source io.Reader) (*Reader, error) {
	reader, err := stream.NewReader(key[:], source)
	if err != nil {
		return nil, err
	}
	return &Reader{
		source:       source,
		notFirst:     false,
		streamReader: reader,
	}, nil
}

func (d *Reader) Read(p []byte) (int, error) {
	if d.notFirst == false {
		if err := d.readFileHeader(); err != nil {
			return 0, err
		}
		d.notFirst = true
	}
	return d.streamReader.Read(p)
}

func (d *Reader) readFileHeader() error {
	err, err2 := d.readFileVersion()
	if err2 != nil {
		return err2
	}

	header, err := d.readHeader()
	if err != nil {
		return err
	}
	d.header = header

	return err
}

func (d *Reader) readFileVersion() (error, error) {
	// Create a buffer to hold the version byte
	versionBuffer := make([]byte, 1)
	_, err := io.ReadFull(d.source, versionBuffer)
	if err != nil {
		return nil, err
	}
	version := versionBuffer[0]

	// Check the version (assuming version 1 is expected)
	if version != 1 {
		return nil, errors.New("unsupported file version")
	}
	return err, nil
}

func (d *Reader) readHeader() (*EncryptedFileHeader, error) {
	headerContext, err := d.readContext()
	if err != nil {
		return nil, err
	}

	// Deserialize file header
	var fileHeader EncryptedFileHeader
	err = json.Unmarshal(headerContext, &fileHeader)
	if err != nil {
		return nil, err
	}

	return &fileHeader, nil
}

func (d *Reader) readContext() ([]byte, error) {
	// Read context size
	contextSize, err := d.readContextSize()
	if err != nil {
		return nil, err
	}

	// Read context
	bufferContext := make([]byte, contextSize)
	n, err := d.source.Read(bufferContext)
	if err != nil {
		return nil, err
	}
	if n != int(contextSize) {
		return nil, errors.New("error reading context")
	}
	return bufferContext, nil
}

func (d *Reader) readContextSize() (uint16, error) {
	var buffer2 [2]byte
	n, err := d.source.Read(buffer2[:])
	if err != nil {
		return 0, err
	}
	if n != 2 {
		return 0, errors.New("error reading context size")
	}
	contextSize := binary.BigEndian.Uint16(buffer2[:])
	return contextSize, nil
}
