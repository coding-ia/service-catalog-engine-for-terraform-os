package main

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	tm "github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"
)

const TestString = "hello world"
const TestBucketHappy = "testBucketHappy"
const TestBucketError = "testBucketError"
const TestObjectPath = "testObjectPath"
const S3ClientErrorMessage = "s3 client error"

type MockDownloader struct {
	mock.Mock
}

func (m *MockDownloader) DownloadObject(ctx context.Context, input *tm.DownloadObjectInput, opts ...func(*tm.Options)) (*tm.DownloadObjectOutput, error) {
	num, _ := input.WriterAt.WriteAt([]byte(TestString), 0)

	if *input.Bucket == TestBucketError {
		return nil, errors.New(S3ClientErrorMessage)
	}

	return &tm.DownloadObjectOutput{
		ContentLength: aws.Int64(int64(num)),
	}, nil
}

func TestS3DownloaderDownloadHappy(t *testing.T) {
	// setup
	downloader := new(MockDownloader)
	s3Downloader := &S3Downloader{
		downloader: downloader,
	}
	expectedResult := []byte(TestString)

	// act
	actualResult, err := s3Downloader.download(t.Context(), TestBucketHappy, TestObjectPath)

	// assert
	if err != nil {
		t.Errorf("Unexpected error occurred")
	}

	if !reflect.DeepEqual(actualResult, expectedResult) {
		t.Errorf("Returned byte array is not the same as expected")
	}
}

func TestS3DownloaderDownloadClientError(t *testing.T) {
	// setup
	downloader := new(MockDownloader)
	s3Downloader := &S3Downloader{
		downloader: downloader,
	}

	// act
	_, err := s3Downloader.download(t.Context(), TestBucketError, TestObjectPath)

	// assert
	if err.Error() != S3ClientErrorMessage {
		t.Errorf("Error is not the same as expected")
	}
}
