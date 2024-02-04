package encryptor

import (
	"cmp"
	"encoding/json"
	"io"
	"slices"
	"sync"
)

const numWorkers = 4

var (
	Separator            = []byte{0, 0, 0, 0} // Define the appropriate separator
	EncryptedFileVersion = []byte{1}          // Define the file version
)

// EncryptReader is a struct for generating encrypted files.
type EncryptReader struct {
	sync.Mutex
	//
	source       io.Reader
	header       EncryptionFileHeader
	key          Key
	buffer       []byte
	nonce        Nonce
	chunkCounter uint32
	chunkSize    uint64
	//
	wg  sync.WaitGroup
	err error
}

// NewEncryptReader creates a new EncryptReader.
func NewEncryptReader(source io.Reader, key Key, chunkSize uint64, clientId string, fileId string, recoveryBlob string) *EncryptReader {
	return &EncryptReader{
		source:       source,
		header:       NewEncryptionFileHeader(chunkSize, clientId, fileId, recoveryBlob),
		buffer:       make([]byte, 0),
		chunkCounter: 0,
		key:          key,
		nonce:        Nonce{},
		chunkSize:    chunkSize,
	}
}

// Read implements the io.Reader interface for EncryptReader.
func (e *EncryptReader) Read(buf []byte) (int, error) {
	if e.chunkCounter == 0 {
		if err := e.appendHeader(); err != nil {
			return 0, err
		}
		e.chunkCounter++
	}

	if len(e.buffer) < len(buf) {
		diff := len(buf) - len(e.buffer)
		err := e.processWithChunkWorkers(int64(diff))
		if err != nil && err != io.EOF {
			return 0, err
		}
	}

	if len(e.buffer) == 0 {
		return 0, io.EOF
	}

	n := copy(buf, e.buffer)
	e.buffer = e.buffer[n:]

	return n, nil // Data read successfully
}

// appendHeader appends the header to the buffer.
func (e *EncryptReader) appendHeader() error {
	e.buffer = append(e.buffer, EncryptedFileVersion...)
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
	e.buffer = append(e.buffer, byte(contextLength), byte(contextLength>>8))
	e.buffer = append(e.buffer, context...)
}

type ChunkData struct {
	Sequence int
	nonce    Nonce
	Data     []byte
}

type ChunkResult struct {
	Sequence int
	Data     []byte
}

type ChunkResultsMut struct {
	sync.Mutex
	list []ChunkResult
}

func (e *EncryptReader) encryptChunks(dc chan ChunkData, resultChan *ChunkResultsMut) {
	defer e.wg.Done()
	for {
		chunkData, ok := <-dc
		if !ok {
			break
		}
		crypto := NewCrypto(e.key, chunkData.nonce)
		encryptedChunk, err := crypto.seal(chunkData.Data)
		if err != nil {
			e.setErr(err)
		}

		resultChan.Lock()
		resultChan.list = append(resultChan.list, ChunkResult{Sequence: chunkData.Sequence, Data: encryptedChunk})
		resultChan.Unlock()
	}
}

// processWithChunkWorkers processes data using concurrent workers and maintains order.
func (e *EncryptReader) processWithChunkWorkers(size int64) error {
	dataChan := make(chan ChunkData, numWorkers)
	results := ChunkResultsMut{}

	// Start workers
	for i := 0; i < numWorkers; i++ {
		e.wg.Add(1)
		go e.encryptChunks(dataChan, &results)
	}

	sequence := 0
	for size > 0 && e.getErr() == nil {
		buffer := make([]byte, e.chunkSize)
		n, err := e.source.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		dataChan <- ChunkData{Sequence: sequence, Data: buffer[:n], nonce: e.nonce}
		e.nonce.increaseBe()
		sequence++
		size -= int64(n)
	}

	close(dataChan)
	e.wg.Wait()

	e.addResultsToBuffer(&results)

	return nil
}

func (e *EncryptReader) addResultsToBuffer(results *ChunkResultsMut) {
	slices.SortFunc(results.list, func(a, b ChunkResult) int {
		return cmp.Compare(a.Sequence, b.Sequence)
	})
	for i := range results.list {
		res := results.list[i]
		e.buffer = append(e.buffer, Separator...)
		e.buffer = append(e.buffer, res.Data...)
	}
}

func (e *EncryptReader) getErr() error {
	e.Lock()
	defer e.Unlock()

	return e.err
}

// setErr is a thread-safe setter for the error object
func (e *EncryptReader) setErr(err error) {
	e.Lock()
	defer e.Unlock()

	e.err = err
}
