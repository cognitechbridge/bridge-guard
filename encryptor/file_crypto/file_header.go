package file_crypto

// fileHeader represents the header of an encryption file
type fileHeader struct {
	Version    string   `json:"version"`
	Alg        string   `json:"alg"`
	ClientID   string   `json:"client_id"`
	FileID     string   `json:"file_id"`
	Recoveries []string `json:"recoveries"`
}

// newEncryptedFileHeader creates a new instance of fileHeader with default values
func newEncryptedFileHeader(clientId string, fileId string, recoveryBlobs []string) fileHeader {
	return fileHeader{
		Version:    "V1",
		Alg:        GetAlgorithmName(), // Set default algorithm
		ClientID:   clientId,
		FileID:     fileId,
		Recoveries: recoveryBlobs,
	}
}
