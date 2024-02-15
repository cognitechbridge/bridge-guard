package types

import "io"

type CloudStorage interface {
	Download(id string, writeAt io.WriterAt) error
	Upload(reader io.Reader, fileId string) error
}
