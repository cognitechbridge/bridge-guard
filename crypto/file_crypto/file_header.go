package file_crypto

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
)

// Header represents the header of an encryption file
type Header struct {
	Version    string   `json:"version"`
	Alg        string   `json:"alg"`
	ClientID   string   `json:"client_id"`
	FileID     string   `json:"file_id"`
	KeyId      string   `json:"key_id"`
	Recoveries []string `json:"recoveries"`
}

// Marshal header
func (h *Header) Marshal() (res []byte, err error) {
	headerBytes, err := json.Marshal(h)
	if err != nil {
		return nil, err
	}
	res, err = formatContext(headerBytes)
	return res, nil
}

// writeContext writes a string to the buffer with its length.
func formatContext(context []byte) (res []byte, err error) {
	contextLength := len(context)
	// Assumes context length fits in 2 bytes
	res = append(res, byte(contextLength>>8), byte(contextLength))
	res = append(res, context...)
	return res, nil
}

func ParseHeader(reader io.Reader) (*Header, error) {
	headerContext, err := readContext(reader)
	if err != nil {
		return nil, err
	}

	// Deserialize file header
	var fileHeader Header
	err = json.Unmarshal(headerContext, &fileHeader)
	if err != nil {
		return nil, err
	}

	return &fileHeader, nil
}

func readContext(reader io.Reader) ([]byte, error) {
	// Read context size
	contextSize, err := readContextSize(reader)
	if err != nil {
		return nil, err
	}

	// Read context
	bufferContext := make([]byte, contextSize)
	n, err := reader.Read(bufferContext)
	if err != nil {
		return nil, err
	}
	if n != int(contextSize) {
		return nil, errors.New("error reading context")
	}
	return bufferContext, nil
}

func readContextSize(reader io.Reader) (uint16, error) {
	var buffer2 [2]byte
	n, err := reader.Read(buffer2[:])
	if err != nil {
		return 0, err
	}
	if n != 2 {
		return 0, errors.New("error reading context size")
	}
	contextSize := binary.BigEndian.Uint16(buffer2[:])
	return contextSize, nil
}
