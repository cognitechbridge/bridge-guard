package persist_file

import "net/http"

const Concurrency = 5

type CtbCloudClient struct {
	baseURL    string
	chunkSize  uint64
	httpClient *http.Client
}

// NewCtbCloudClient NewUploaderClient creates a new CtbCloudClient.
func NewCtbCloudClient(baseURL string, chunkSize uint64) *CtbCloudClient {
	return &CtbCloudClient{
		baseURL:    baseURL,
		chunkSize:  chunkSize,
		httpClient: &http.Client{},
	}
}
