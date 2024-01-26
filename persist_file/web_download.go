package persist_file

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

type downloader struct {
	sync.Mutex
	writeAt   io.WriterAt
	fileName  string
	pos       int64
	wg        sync.WaitGroup
	err       error
	client    *CtbCloudClient
	chunkSize uint64

	totalBytes int64
}

type dlchunk struct {
	w       io.WriterAt
	start   int64
	size    int64
	cur     int64
	partNum int
}

func (c *CtbCloudClient) Download(fileName string, writeAt io.WriterAt) error {
	d := downloader{
		fileName:  fileName,
		wg:        sync.WaitGroup{},
		writeAt:   writeAt,
		chunkSize: c.ChunkSize,
		client:    c,
	}
	return d.Download()
}

func (d *downloader) Download() error {

	// Spin off first worker to check additional header information
	d.getChunk()

	total := d.getTotalBytes()
	if total <= 0 {
		return fmt.Errorf("cannot read total parts")
	}

	// Spin up workers
	ch := make(chan dlchunk, Concurrency)

	for i := 0; i < Concurrency; i++ {
		d.wg.Add(1)
		go d.downloadPart(ch)
	}

	// Assign work
	for d.getErr() == nil {
		if d.pos >= total {
			break // We're finished queuing chunks
		}

		// Queue the next range of bytes to read.
		ch <- dlchunk{w: d.writeAt, start: d.pos, size: int64(d.chunkSize)}
		d.pos += int64(d.chunkSize)
	}

	// Wait for completion
	close(ch)
	d.wg.Wait()

	return nil
}

// downloadPart is an individual goroutine worker reading from the ch channel
// and performing a GetObject request on the data with a given byte range.
//
// If this is the first worker, this operation also resolves the total number
// of bytes to be read so that the worker manager knows when it is finished.
func (d *downloader) downloadPart(ch chan dlchunk) {
	defer d.wg.Done()
	for {
		chunk, ok := <-ch
		if !ok {
			break
		}
		if d.getErr() != nil {
			// Drain the channel if there is an error, to prevent deadlocking
			// of download producer.
			continue
		}

		if err := d.downloadChunk(chunk); err != nil {
			d.setErr(err)
		}
	}
}

// getChunk grabs a chunk of data from the body.
// Not thread safe. Should only used when grabbing data on a single thread.
func (d *downloader) getChunk() {
	if d.getErr() != nil {
		return
	}

	chunk := dlchunk{w: d.writeAt, start: d.pos, size: int64(d.client.ChunkSize)}
	d.pos += int64(d.client.ChunkSize)

	if err := d.downloadChunk(chunk); err != nil {
		d.setErr(err)
	}
}

// downloadChunk downloads the chunk from s3
func (d *downloader) downloadChunk(chunk dlchunk) error {
	query := url.Values{}
	query.Add("start", fmt.Sprintf("%d", chunk.start))
	query.Add("size", fmt.Sprintf("%d", chunk.size))
	reqURL := fmt.Sprintf(
		"%s/download/%s?%s",
		d.client.baseURL,
		url.PathEscape(d.fileName),
		query.Encode(),
	)
	req, err := http.NewRequest("POST", reqURL, nil)

	client := http.Client{}
	reqResponse, err := client.Do(req)
	if err != nil {
		return err
	}
	defer reqResponse.Body.Close()

	if err != nil {
		return err
	}

	// Read data into the buffer.
	buf := make([]byte, chunk.size)
	bytesRead, readErr := reqResponse.Body.Read(buf)
	if readErr != nil && readErr != io.EOF {
		return readErr
	}
	if bytesRead > 0 {
		// Write data at the specific offset.
		_, writeErr := d.writeAt.WriteAt(buf[:bytesRead], chunk.start)
		if writeErr != nil {
			panic(writeErr)
		}
	}

	totalBytesStr := reqResponse.Header.Get("Total-Bytes")
	d.setTotalBytes(totalBytesStr)

	return err
}

func (d *downloader) getTotalBytes() int64 {
	d.Lock()
	defer d.Unlock()

	return d.totalBytes
}

func (d *downloader) setTotalBytes(totalBytesStr string) {
	d.Lock()
	defer d.Unlock()

	if d.totalBytes > 0 {
		return
	}

	totalBytes, _ := strconv.ParseInt(totalBytesStr, 10, 64)
	d.totalBytes = totalBytes
}

// getErr is a thread-safe getter for the error object
func (d *downloader) getErr() error {
	d.Lock()
	defer d.Unlock()

	return d.err
}

// setErr is a thread-safe setter for the error object
func (d *downloader) setErr(e error) {
	d.Lock()
	defer d.Unlock()

	d.err = e
}
