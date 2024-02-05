package objectstorage

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
)

type DummyClient struct {
	BucketName string
	ChunkSize  int64
	Client     *s3.Client
}

func NewDummyClient() *DummyClient {
	return &DummyClient{}
}

func (s *DummyClient) Upload(reader io.Reader, key string) error {
	buf := make([]byte, 10*1024*1024)
	i := 0
	for {
		_, err := reader.Read(buf)
		if err != nil {
			break
		}
		i++
	}
	return nil
}

func (s *DummyClient) Download(key string, writeAt io.WriterAt) error {
	return nil
}
