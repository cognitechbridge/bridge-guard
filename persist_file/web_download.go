package persist_file

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

func (c *CtbCloudClient) Download(key string, writeAt io.WriterAt) error {
	offset := uint64(0)
	partNum := 1
	for {
		query := url.Values{}
		query.Add("partnumber", fmt.Sprintf("%d", partNum))
		reqURL := fmt.Sprintf(
			"%s/download/%s?%s",
			c.baseURL,
			url.PathEscape(key),
			query.Encode(),
		)
		req, err := http.NewRequest("POST", reqURL, nil)

		reqResponse, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer reqResponse.Body.Close()

		if err != nil {
			return err
		}

		// Read data into the buffer.
		buf := make([]byte, c.ChunkSize)
		bytesRead, readErr := reqResponse.Body.Read(buf)
		if readErr != nil && readErr != io.EOF {
			return readErr
		}
		if bytesRead > 0 {
			// Write data at the specific offset.
			_, writeErr := writeAt.WriteAt(buf[:bytesRead], int64(offset))
			if writeErr != nil {
				panic(writeErr)
			}
			// Update the offset.
			offset += uint64(bytesRead)
			partNum += 1
		}

		partsCountStr := reqResponse.Header.Get("Parts-Count")
		partsCount, _ := strconv.Atoi(partsCountStr)

		if partNum > partsCount {
			break
		}

	}

	return nil
}
