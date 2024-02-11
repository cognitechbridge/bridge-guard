package encryptor

const DefaultChunkSize = 1024

// EncryptedFileHeader represents the header of an encryption file
type EncryptedFileHeader struct {
	Version    string   `json:"version"`
	Alg        string   `json:"alg"`
	ClientID   string   `json:"client_id"`
	FileID     string   `json:"file_id"`
	Recoveries []string `json:"recoveries"`
}

// NewEncryptedFileHeader creates a new instance of EncryptedFileHeader with default values
func NewEncryptedFileHeader(clientId string, fileId string, recoveryBlobs []string) EncryptedFileHeader {
	return EncryptedFileHeader{
		Version:    "V1",
		Alg:        GetAlgorithmName(), // Set default algorithm
		ClientID:   clientId,
		FileID:     fileId,
		Recoveries: recoveryBlobs,
	}
}
