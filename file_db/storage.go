package file_db

import "io"

type CloudStorageClient interface {
	Upload(reader io.Reader, fileId string, fileName string) error
	Download(fileId string, writeAt io.WriterAt) error
}
