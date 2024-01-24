package encryptor

const DefaultChunkSize = 1024

// EncryptionFileHeader represents the header of an encryption file
type EncryptionFileHeader struct {
	Version   string `json:"version"`
	Alg       string `json:"alg"`
	ClientID  string `json:"client_id"`
	FileID    string `json:"file_id"`
	ChunkSize uint64 `json:"chunk_size"`
	Recovery  string `json:"recovery"`
}

// NewEncryptionFileHeader creates a new instance of EncryptionFileHeader with default values
func NewEncryptionFileHeader(chunkSize uint64, clientId string, fileId string, recoverBlob string) EncryptionFileHeader {
	return EncryptionFileHeader{
		Version:   "V1",
		ChunkSize: chunkSize,          // Define this constant as per your requirements
		Alg:       GetAlgorithmName(), // Set default algorithm
		ClientID:  clientId,
		FileID:    fileId,
		Recovery:  recoverBlob,
	}
}
