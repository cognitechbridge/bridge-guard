package encryptor

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
)

type DecryptReader struct {
	source       io.Reader
	key          *Key
	nonce        Nonce
	buffer       []byte
	chunkSize    uint64
	chunkCounter uint64
}

func NewDecryptReader(key *Key, source io.Reader) (*DecryptReader, error) {

	return &DecryptReader{
		source: source,
		key:    key,
		nonce:  Nonce{},
		buffer: make([]byte, 0),
	}, nil
}

func (d *DecryptReader) Read(p []byte) (int, error) {
	if d.chunkSize == 0 {
		if err := d.readFileHeader(); err != nil {
			return 0, err
		}
	}

	for len(d.buffer) < len(p) {
		err := d.readChunkHeader()
		if err == io.EOF {
			break
		} else if err != nil {
			return 0, err
		}

		buffer := make([]byte, d.chunkSize)
		bytesRead, err := d.source.Read(buffer)
		if err != nil {
			return 0, err
		}
		if bytesRead == 0 {
			break
		}

		crypto := NewCrypto(*d.key, d.nonce)
		decryptedData, err := crypto.open(buffer[:bytesRead])
		if err != nil {
			return 0, err
		}
		d.chunkCounter++

		d.buffer = append(d.buffer, decryptedData...)
		d.nonce.increaseBe()
	}

	if len(d.buffer) == 0 {
		return 0, io.EOF
	}

	n := copy(p, d.buffer)
	d.buffer = d.buffer[n:]
	return n, nil
}

func (d *DecryptReader) readFileHeader() error {
	err, err2 := d.readFileVersion()
	if err2 != nil {
		return err2
	}

	header, err := d.readHeader()
	if err != nil {
		return err
	}
	d.chunkSize = header.ChunkSize + uint64(GetOverHeadSize())

	return err
}

func (d *DecryptReader) readFileVersion() (error, error) {
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

func (d *DecryptReader) readChunkHeader() error {
	// Define a small buffer for the chunk header (4 bytes as in Rust code)
	var smallBuffer [4]byte

	// Read 4 bytes into the buffer
	n, err := d.source.Read(smallBuffer[:])
	if err != nil {
		return err
	}

	// If no bytes were read, return nil (end of stream or chunk)
	if n == 0 {
		return nil
	}

	// Check if all bytes are zero, which is a specific condition
	if n == 4 && isZeroes(smallBuffer[:]) {
		return nil
	}

	// If the header is not valid, return an error
	return errors.New("chunk header is not valid")
}

func (d *DecryptReader) readHeader() (*EncryptedFileHeader, error) {
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

func (d *DecryptReader) readContext() ([]byte, error) {
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

func (d *DecryptReader) readContextSize() (uint16, error) {
	var buffer2 [2]byte
	n, err := d.source.Read(buffer2[:])
	if err != nil {
		return 0, err
	}
	if n != 2 {
		return 0, errors.New("error reading context size")
	}
	contextSize := binary.LittleEndian.Uint16(buffer2[:])
	return contextSize, nil
}
