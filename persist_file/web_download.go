package persist_file

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

const Concurrency = 5

type downloader struct {
	sync.Mutex
	writeAt  io.WriterAt
	fileName string
	pos      int64
	part     int
	wg       sync.WaitGroup
	err      error
	client   *CtbCloudClient

	totalParts int
}

type dlchunk struct {
	w       io.WriterAt
	start   int64
	size    int64
	cur     int64
	partNum int
}

func (c *CtbCloudClient) Download(writeAt io.WriterAt, fileName string) error {
	d := downloader{
		fileName: fileName,
		wg:       sync.WaitGroup{},
		writeAt:  writeAt,
		client:   c,
	}
	return d.Download()
}

func (d *downloader) Download() error {

	// Spin off first worker to check additional header information
	d.getChunk()

	total := d.getTotalParts()
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
		if d.part >= total {
			break // We're finished queuing chunks
		}

		// Queue the next range of bytes to read.
		ch <- dlchunk{w: d.writeAt, start: d.pos, size: int64(d.client.ChunkSize)}
		d.pos += int64(d.client.ChunkSize)
	}

	// Wait for completion
	close(ch)
	d.wg.Wait()

	return nil
}

// getChunk grabs a chunk of data from the body.
// Not thread safe. Should only used when grabbing data on a single thread.
func (d *downloader) getChunk() {
	if d.getErr() != nil {
		return
	}

	chunk := dlchunk{w: d.writeAt, start: d.pos, size: int64(d.client.ChunkSize)}
	d.pos += int64(d.client.ChunkSize)
	d.part += 1

	if err := d.downloadChunk(chunk); err != nil {
		d.setErr(err)
	}
}

// downloadChunk downloads the chunk from s3
func (d *downloader) downloadChunk(chunk dlchunk) error {
	query := url.Values{}
	query.Add("partnumber", fmt.Sprintf("%d", chunk.partNum))
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

	partsCountStr := reqResponse.Header.Get("Parts-Count")
	d.setTotalParts(partsCountStr)

	return err
}

func (d *downloader) getTotalParts() int {
	d.Lock()
	defer d.Unlock()

	return d.totalParts
}

func (d *downloader) setTotalParts(partsCountStr string) {
	d.Lock()
	defer d.Unlock()

	if d.totalParts >= 0 {
		return
	}

	totalParts, _ := strconv.Atoi(partsCountStr)
	d.totalParts = totalParts
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
