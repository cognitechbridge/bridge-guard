package encryptor

import (
	"ctb-cli/encryptor/stream"
	"encoding/json"
	"io"
)

var (
	EncryptedFileVersion = []byte{1} // Define the file version
)

// Writer is a struct for generating encrypted files.
type Writer struct {
	header       EncryptedFileHeader
	notFirst     bool
	dst          io.Writer
	streamWriter *stream.Writer
}

// NewWriter creates a new Writer.
func NewWriter(dst io.Writer, key Key, chunkSize uint64, clientId string, fileId string, recoveryBlobs []string) (*Writer, error) {
	writer, err := stream.NewWriter(key[:], dst)
	if err != nil {
		return nil, err
	}
	return &Writer{
		dst:          dst,
		header:       NewEncryptedFileHeader(chunkSize, clientId, fileId, recoveryBlobs),
		notFirst:     false,
		streamWriter: writer,
	}, nil
}

func (e *Writer) Write(buf []byte) (int, error) {
	if e.notFirst == false {
		if err := e.appendHeader(); err != nil {
			return 0, err
		}
		e.notFirst = true
	}
	return e.streamWriter.Write(buf)
}

// appendHeader appends the header to the buffer.
func (e *Writer) appendHeader() (err error) {
	_, err = e.dst.Write(EncryptedFileVersion)
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
func (e *Writer) writeContext(context string) (err error) {
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

func (e *Writer) Close() error {
	return e.streamWriter.Close()
}
