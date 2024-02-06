package encryptor

const DefaultChunkSize = 1024

// EncryptedFileHeader represents the header of an encryption file
type EncryptedFileHeader struct {
	Version    string   `json:"version"`
	Alg        string   `json:"alg"`
	ClientID   string   `json:"client_id"`
	FileID     string   `json:"file_id"`
	ChunkSize  uint64   `json:"chunk_size"`
	Recoveries []string `json:"recoveries"`
}

// NewEncryptedFileHeader creates a new instance of EncryptedFileHeader with default values
func NewEncryptedFileHeader(chunkSize uint64, clientId string, fileId string, recoveryBlobs []string) EncryptedFileHeader {
	return EncryptedFileHeader{
		Version:    "V1",
		ChunkSize:  chunkSize,          // Define this constant as per your requirements
		Alg:        GetAlgorithmName(), // Set default algorithm
		ClientID:   clientId,
		FileID:     fileId,
		Recoveries: recoveryBlobs,
	}
}
