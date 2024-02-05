package s3

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"log"
)

// Client S3Client represents the objectstorage configuration for S3
type Client struct {
	BucketName string
	ChunkSize  int64
	Client     *s3.Client
}

// NewClient creates a new instance of S3Client
func NewClient(bucketName string, chunkSize int64) *Client {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)
	return &Client{
		BucketName: bucketName,
		ChunkSize:  chunkSize,
		Client:     client,
	}
}

func (s *Client) Upload(reader io.Reader, key string) error {
	uploader := manager.NewUploader(s.Client, func(u *manager.Uploader) {
		u.PartSize = s.ChunkSize
	})
	_, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
		Body:   reader,
	})
	if err != nil {
		return fmt.Errorf("Couldn't upload object. Here's why: %v\n", err)
	}
	return err
}

func (s *Client) Download(key string, writeAt io.WriterAt) error {
	var partMiBs int64 = 10
	downloader := manager.NewDownloader(s.Client, func(d *manager.Downloader) {
		d.PartSize = partMiBs * 1024 * 1024
	})
	_, err := downloader.Download(context.TODO(), writeAt, &s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("Couldn't download. Here's why: %v\n", err)
	}
	return nil
}
