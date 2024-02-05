package cloud

import "net/http"

const Concurrency = 5

type Client struct {
	baseURL    string
	chunkSize  uint64
	httpClient *http.Client
}

// NewClient NewUploaderClient creates a new Client.
func NewClient(baseURL string, chunkSize uint64) *Client {
	return &Client{
		baseURL:    baseURL,
		chunkSize:  chunkSize,
		httpClient: &http.Client{},
	}
}
