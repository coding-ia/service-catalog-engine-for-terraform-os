package main

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	tm "github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"
)

const TestArtifactPath = "s3://terraform-configurations-cross-account-demo/product_with_override_var.tar.gz"
const TestArtifactType = "AWS_S3"
const TestLaunchRoleArn = "arn:aws:iam::829064435212:role/SCLaunchRole"
const TestS3BucketArtifactPath = "../../sample-provisioning-artifacts/s3bucket.tar.gz"
const TestS3BucketArtifactFileName = "main.tf"
const TestS3BucketArtifactFileContent = "\"bucket_name\" {\n  type = string\n}\nprovider \"aws\" {\n}\nresource \"aws_s3_bucket\" \"bucket\" {\n  bucket = var.bucket_name\n}\noutput regional_domain_name {\n  value = aws_s3_bucket.bucket.bucket_regional_domain_name\n}"

type MockS3Downloader struct {
	mock.Mock
}

func (m *MockS3Downloader) DownloadObject(ctx context.Context, input *tm.DownloadObjectInput, opts ...func(*tm.Options)) (*tm.DownloadObjectOutput, error) {
	if *input.Bucket != "terraform-configurations-cross-account-demo" || *input.Key != "product_with_override_var.tar.gz" {
		return nil, errors.New(S3ClientErrorMessage)
	}

	b, _ := os.ReadFile(TestS3BucketArtifactPath)
	numBytes, _ := input.WriterAt.WriteAt(b, 0)

	return &tm.DownloadObjectOutput{
		ContentLength: aws.Int64(int64(numBytes)),
	}, nil
}

func TestConfigFetcherFetchHappy(t *testing.T) {
	// setup
	downloader := new(MockS3Downloader)
	s3Downloader := &S3Downloader{
		downloader: downloader,
	}
	configFetcher := &ConfigFetcher{
		s3Downloader: s3Downloader,
	}
	input := TerraformOpenSourceParameterParserInput{
		Artifact: Artifact{
			Path: TestArtifactPath,
			Type: TestArtifactType,
		},
		LaunchRoleArn: TestLaunchRoleArn,
	}

	// act
	fileMap, err := configFetcher.fetch(t.Context(), input)

	// assert
	if err != nil {
		t.Errorf("Unexpected error occured")
	}

	fileContent, ok := fileMap[TestS3BucketArtifactFileName]
	if !ok {
		t.Errorf("Expected file %s was not parsed", TestS3BucketArtifactFileName)
	}

	if reflect.DeepEqual(fileContent, TestS3BucketArtifactFileContent) {
		t.Errorf("File content for %s is not as expected", TestS3BucketArtifactFileName)
	}
}

func TestConfigFetcherFetchWithEmptyLaunchRoleHappy(t *testing.T) {
	// setup
	downloader := new(MockS3Downloader)
	s3Downloader := &S3Downloader{
		downloader: downloader,
	}
	configFetcher := &ConfigFetcher{
		s3Downloader: s3Downloader,
	}
	input := TerraformOpenSourceParameterParserInput{
		Artifact: Artifact{
			Path: TestArtifactPath,
			Type: TestArtifactType,
		},
		LaunchRoleArn: "",
	}

	// act
	fileMap, err := configFetcher.fetch(t.Context(), input)

	// assert
	if err != nil {
		t.Errorf("Unexpected error occured")
	}

	fileContent, ok := fileMap[TestS3BucketArtifactFileName]
	if !ok {
		t.Errorf("Expected file %s was not parsed", TestS3BucketArtifactFileName)
	}

	if reflect.DeepEqual(fileContent, TestS3BucketArtifactFileContent) {
		t.Errorf("File content for %s is not as expected", TestS3BucketArtifactFileName)
	}
}
