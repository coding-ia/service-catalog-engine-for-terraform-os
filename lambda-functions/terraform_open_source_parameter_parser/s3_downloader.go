package main

import (
	"context"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	tm "github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Downloader interface {
	DownloadObject(ctx context.Context, input *tm.DownloadObjectInput, opts ...func(*tm.Options)) (*tm.DownloadObjectOutput, error)
}

type S3Downloader struct {
	downloader Downloader
}

func NewS3Downloader(ctx context.Context, credsProvider aws.CredentialsProvider) (*S3Downloader, error) {
	// load default config, overriding credentials with the provided provider
	cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(credsProvider))
	if err != nil {
		return &S3Downloader{}, err
	}

	// create an S3 client from the config
	s3Client := s3.NewFromConfig(cfg)

	// create a new Transfer Manager client using the S3 client
	return &S3Downloader{downloader: tm.New(s3Client)}, nil
}

// Downloads artifact from specified S3 bucket and objectPath in a byte array
func (client *S3Downloader) download(ctx context.Context, bucket string, objectPath string) ([]byte, error) {
	// open byte array as download target
	buff := &writeAtBuffer{}

	// download to buffer
	log.Printf("Downloading %s from bucket %s", objectPath, bucket)
	_, err := client.downloader.DownloadObject(ctx, &tm.DownloadObjectInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(objectPath),
		WriterAt: buff,
	})
	if err != nil {
		return []byte{}, err
	}

	numBytes := len(buff.Bytes())
	log.Printf("Downloaded %d bytes", numBytes)
	return buff.Bytes(), nil
}

// writeAtBuffer is a minimal, concurrency-safe io.WriterAt backed by an
// in-memory byte slice, used as a download target. Replaces the deprecated
// manager.WriteAtBuffer from feature/s3/manager.
type writeAtBuffer struct {
	mu  sync.Mutex
	buf []byte
}

func (b *writeAtBuffer) WriteAt(p []byte, off int64) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	end := off + int64(len(p))
	if end > int64(len(b.buf)) {
		newBuf := make([]byte, end)
		copy(newBuf, b.buf)
		b.buf = newBuf
	}
	copy(b.buf[off:], p)
	return len(p), nil
}

func (b *writeAtBuffer) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf
}
