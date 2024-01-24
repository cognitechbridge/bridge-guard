package storage

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

// S3Storage represents the storage configuration for S3
type S3Storage struct {
	BucketName string
	ChunkSize  int64
	Client     *s3.Client
}

// NewS3Storage creates a new instance of S3Storage
func NewS3Storage(bucketName string, chunkSize int64) *S3Storage {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)
	return &S3Storage{
		BucketName: bucketName,
		ChunkSize:  chunkSize,
		Client:     client,
	}
}

func (s *S3Storage) Upload(reader io.Reader, key string) error {
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

func (s *S3Storage) Download(key string, writeAt io.WriterAt) error {
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
