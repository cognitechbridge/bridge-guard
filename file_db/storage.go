package file_db

import "io"

type CloudStorageClient interface {
	Upload(reader io.Reader, fileName string) error
	Download(key string, writeAt io.WriterAt) error
}
