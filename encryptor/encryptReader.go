package encryptor

import (
	"ctb-cli/encryptor/stream"
	"encoding/json"
	"io"
	"sync"
)

const numWorkers = 4

var (
	EncryptedFileVersion = []byte{1} // Define the file version
)

// EncryptReader is a struct for generating encrypted files.
type EncryptReader struct {
	sync.Mutex
	//
	header       EncryptedFileHeader
	chunkCounter uint32
	//
	dir io.Writer
	//
	writer *stream.Writer
}

// NewEncryptReader creates a new EncryptReader.
func NewEncryptReader(source io.Writer, key Key, chunkSize uint64, clientId string, fileId string, recoveryBlobs []string) *EncryptReader {
	writer, _ := stream.NewWriter(key[:], source)
	return &EncryptReader{
		dir:          source,
		header:       NewEncryptedFileHeader(chunkSize, clientId, fileId, recoveryBlobs),
		chunkCounter: 0,
		writer:       writer,
	}
}

func (e *EncryptReader) Write(buf []byte) (int, error) {
	if e.chunkCounter == 0 {
		if err := e.appendHeader(); err != nil {
			return 0, err
		}
		e.chunkCounter++
	}
	return e.writer.Write(buf)
}

// appendHeader appends the header to the buffer.
func (e *EncryptReader) appendHeader() error {
	_, err := e.dir.Write(EncryptedFileVersion)
	if err != nil {
		return err
	}
	headerBytes, err := json.Marshal(e.header)
	if err != nil {
		return err
	}
	e.writeContext(string(headerBytes))
	return nil
}

// writeContext writes a string to the buffer with its length.
func (e *EncryptReader) writeContext(context string) {
	contextLength := len(context)
	// Assumes context length fits in 2 bytes
	_, _ = e.dir.Write([]byte{byte(contextLength), byte(contextLength >> 8)})
	_, _ = e.dir.Write([]byte(context))
}

func (e *EncryptReader) Close() error {
	return e.writer.Close()
}
