package persist_file

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
)

const Concurrency = 5

type Uploader struct {
	sync.Mutex
	reader   io.Reader
	fileName string
	wg       sync.WaitGroup
	err      error
	client   *CtbCloudClient
}

type chunk struct {
	buf []byte
	num int32
}

func (c *CtbCloudClient) Upload(reader io.Reader, fileName string) error {
	u := Uploader{
		fileName: fileName,
		wg:       sync.WaitGroup{},
		reader:   reader,
		client:   c,
	}
	return u.Upload()
}

func (u *Uploader) Upload() error {
	partNumber := int32(1)

	ch := make(chan chunk, Concurrency)
	for i := 0; i < Concurrency; i++ {
		u.wg.Add(1)
		go u.readChunk(ch)
	}

	for u.geterr() == nil {
		// Create a buffer to store multipart form data
		buf := make([]byte, u.client.ChunkSize)
		written, err := u.reader.Read(buf)

		// Check for errors other than EOF
		if err != nil && err != io.EOF {
			return err
		}
		if err != nil && err == io.EOF {
			break
		}

		ch <- chunk{buf: buf[:written], num: partNumber}

		// Increment the part number
		partNumber++
	}

	// Close the channel, wait for workers, and complete upload
	close(ch)
	u.wg.Wait()

	// After uploading all parts, send a request to `/upload/complete` with query parameter
	err := u.finishUpload()
	if err != nil {
		return err
	}

	return nil
}

func (u *Uploader) finishUpload() error {
	reqURL := fmt.Sprintf(
		"%s/upload/%s/complete",
		u.client.baseURL,
		url.PathEscape(u.fileName),
	)
	completeReq, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		return err
	}

	// Send the request
	httpClient := http.Client{}
	completeResponse, err := httpClient.Do(completeReq)
	if err != nil {
		return err
	}
	defer completeResponse.Body.Close()

	if completeResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("upload completion failed with status code: %d", completeResponse.StatusCode)
	}
	return nil
}

func (u *Uploader) readChunk(ch chan chunk) {
	defer u.wg.Done()
	for {
		data, ok := <-ch

		if !ok {
			break
		}

		if u.geterr() == nil {
			if err := u.send(data); err != nil {
				u.seterr(err)
			}
		}
	}
}

func (u *Uploader) send(ch chunk) error {
	// Create and send the request with query parameters
	query := url.Values{}
	query.Add("partnumber", fmt.Sprintf("%d", ch.num))
	reqURL := fmt.Sprintf(
		"%s/upload/%s?%s",
		u.client.baseURL,
		url.PathEscape(u.fileName),
		query.Encode(),
	)

	buf := bytes.NewBuffer(ch.buf)
	req, err := http.NewRequest("POST", reqURL, buf)
	if err != nil {
		return err
	}

	// Send the request
	client := http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	_ = response.Body.Close()

	return nil
}

// geterr is a thread-safe getter for the error object
func (u *Uploader) geterr() error {
	u.Lock()
	defer u.Unlock()

	return u.err
}

// seterr is a thread-safe setter for the error object
func (u *Uploader) seterr(e error) {
	u.Lock()
	defer u.Unlock()

	u.err = e
}
