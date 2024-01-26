package persist_file

import "net/http"

const Concurrency = 5

type CtbCloudClient struct {
	baseURL    string
	ChunkSize  uint64
	httpClient *http.Client
}

// NewCtbCloudClient NewUploaderClient creates a new CtbCloudClient.
func NewCtbCloudClient(baseURL string, chunkSize uint64) *CtbCloudClient {
	return &CtbCloudClient{
		baseURL:    baseURL,
		ChunkSize:  chunkSize,
		httpClient: &http.Client{},
	}
}
