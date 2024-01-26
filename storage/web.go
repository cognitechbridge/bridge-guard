package storage

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type CtbCloudClient struct {
	baseURL    string
	ChunkSize  int64
	httpClient *http.Client
}

// NewUploaderClient creates a new CtbCloudClient.
func NewUploaderClient(baseURL string, chunkSize int64) *CtbCloudClient {
	return &CtbCloudClient{
		baseURL:    baseURL,
		ChunkSize:  chunkSize,
		httpClient: &http.Client{},
	}
}

func (client *CtbCloudClient) Upload(reader io.Reader, fileName string) error {
	partNumber := int64(1)

	for {
		// Create a buffer to store multipart form data
		buf := make([]byte, client.ChunkSize)
		written, err := reader.Read(buf)
		buffer := bytes.NewBuffer(buf[:written])

		// Check for errors other than EOF
		if err != nil && err != io.EOF {
			return err
		}
		if err != nil && err == io.EOF {
			break
		}

		// Create and send the request with query parameters
		query := url.Values{}
		query.Add("filename", fileName)
		query.Add("partnumber", fmt.Sprintf("%d", partNumber))
		reqURL := client.baseURL + "/upload?" + query.Encode()

		req, err := http.NewRequest("POST", reqURL, buffer)
		if err != nil {
			return err
		}

		// Send the request
		response, err := client.httpClient.Do(req)
		if err != nil {
			return err
		}
		_ = response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("upload failed with status code: %d", response.StatusCode)
		}

		// Increment the part number
		partNumber++
	}

	// After uploading all parts, send a request to `/upload/complete` with query parameter

	query := url.Values{}
	query.Add("filename", fileName)
	reqURL := client.baseURL + "/upload/complete?" + query.Encode()
	completeReq, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		return err
	}

	// Send the request
	completeResponse, err := client.httpClient.Do(completeReq)
	if err != nil {
		return err
	}
	defer completeResponse.Body.Close()

	if completeResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("upload completion failed with status code: %d", completeResponse.StatusCode)
	}

	return nil
}

func (client *CtbCloudClient) Download(key string, writeAt io.WriterAt) error {
	query := url.Values{}
	query.Add("filename", key)
	reqURL := client.baseURL + "/download?" + query.Encode()
	req, err := http.NewRequest("POST", reqURL, nil)

	reqResponse, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer reqResponse.Body.Close()

	if err != nil {
		return err
	}

	offset := int64(0)
	buf := make([]byte, client.ChunkSize)
	for {
		// Read data into the buffer.
		bytesRead, readErr := reqResponse.Body.Read(buf)
		if bytesRead > 0 {
			// Write data at the specific offset.
			_, writeErr := writeAt.WriteAt(buf[:bytesRead], offset)
			if writeErr != nil {
				panic(writeErr)
			}
			// Update the offset.
			offset += int64(bytesRead)
		}

		if readErr == io.EOF {
			break // End of file reached
		}
		if readErr != nil {
			return readErr
		}
	}

	return nil
}
