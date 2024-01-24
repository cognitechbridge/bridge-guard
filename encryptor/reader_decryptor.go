package encryptor

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
)

type ReaderDecryptor struct {
	source       io.Reader
	key          Key
	nonce        Nonce
	buffer       []byte
	chunkSize    uint64
	chunkCounter uint64
}

func NewReaderDecryptor(key Key, source io.Reader) (*ReaderDecryptor, error) {

	return &ReaderDecryptor{
		source: source,
		key:    key,
		nonce:  Nonce{},
		buffer: make([]byte, 0),
	}, nil
}

func (rd *ReaderDecryptor) Read(p []byte) (int, error) {
	if rd.chunkSize == 0 {
		if err := rd.readFileHeader(); err != nil {
			return 0, err
		}
	}

	for len(rd.buffer) < len(p) {
		err := rd.readChunkHeader()
		if err == io.EOF {
			break
		} else if err != nil {
			return 0, err
		}

		buffer := make([]byte, rd.chunkSize)
		bytesRead, err := rd.source.Read(buffer)
		if err != nil {
			return 0, err
		}
		if bytesRead == 0 {
			break
		}

		decryptedData, err := DecryptChunk(buffer[:bytesRead], rd.key, rd.nonce)
		if err != nil {
			return 0, err
		}
		rd.chunkCounter++

		rd.buffer = append(rd.buffer, decryptedData...)
		rd.nonce.increaseBe()
	}

	if len(rd.buffer) == 0 {
		return 0, io.EOF
	}

	n := copy(p, rd.buffer)
	rd.buffer = rd.buffer[n:]
	return n, nil
}

func (rd *ReaderDecryptor) readFileHeader() error {
	err, err2 := rd.readFileVersion()
	if err2 != nil {
		return err2
	}

	header, err := rd.readHeader()
	if err != nil {
		return err
	}
	rd.chunkSize = header.ChunkSize + uint64(GetOverHeadSize())

	return err
}

func (rd *ReaderDecryptor) readFileVersion() (error, error) {
	// Create a buffer to hold the version byte
	versionBuffer := make([]byte, 1)
	_, err := io.ReadFull(rd.source, versionBuffer)
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

func (rd *ReaderDecryptor) readChunkHeader() error {
	// Define a small buffer for the chunk header (4 bytes as in Rust code)
	var smallBuffer [4]byte

	// Read 4 bytes into the buffer
	n, err := rd.source.Read(smallBuffer[:])
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

func (rd *ReaderDecryptor) readHeader() (*EncryptionFileHeader, error) {
	headerContext, err := rd.readContext()
	if err != nil {
		return nil, err
	}

	// Deserialize file header
	var fileHeader EncryptionFileHeader
	err = json.Unmarshal(headerContext, &fileHeader)
	if err != nil {
		return nil, err
	}

	return &fileHeader, nil
}

func (rd *ReaderDecryptor) readContext() ([]byte, error) {
	// Read context size
	contextSize, err := rd.readContextSize()
	if err != nil {
		return nil, err
	}

	// Read context
	bufferContext := make([]byte, contextSize)
	n, err := rd.source.Read(bufferContext)
	if err != nil {
		return nil, err
	}
	if n != int(contextSize) {
		return nil, errors.New("error reading context")
	}
	return bufferContext, nil
}

func (rd *ReaderDecryptor) readContextSize() (uint16, error) {
	var buffer2 [2]byte
	n, err := rd.source.Read(buffer2[:])
	if err != nil {
		return 0, err
	}
	if n != 2 {
		return 0, errors.New("error reading context size")
	}
	contextSize := binary.LittleEndian.Uint16(buffer2[:])
	return contextSize, nil
}
