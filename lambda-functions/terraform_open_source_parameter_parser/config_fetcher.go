package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

const LaunchRoleAccessDeniedErrorMessage = "Access denied while assuming launch role %s: %s"
const ArtifactFetchAccessDeniedErrorMessage = "Access denied while downloading artifact from %s: %s"
const UnzipFailureErrorMessage = "Artifact from %s is not a valid tar.gz file: %s"

type ConfigFetcher struct {
	s3Downloader *S3Downloader
}

func NewConfigFetcher(ctx context.Context, launchRoleArn string) (*ConfigFetcher, error) {
	credsProvider, err := retrieveConfigFetcherCreds(ctx, launchRoleArn)
	if err != nil {
		return &ConfigFetcher{},
			ParserAccessDeniedException{Message: fmt.Sprintf(LaunchRoleAccessDeniedErrorMessage, launchRoleArn, err.Error())}
	}

	s3Downloader, err := NewS3Downloader(ctx, credsProvider)
	if err != nil {
		return &ConfigFetcher{},
			ParserAccessDeniedException{Message: fmt.Sprintf(LaunchRoleAccessDeniedErrorMessage, launchRoleArn, err.Error())}
	}

	return &ConfigFetcher{s3Downloader: s3Downloader}, nil
}

// Fetches the input file from artifact location and outputs as a map of the file's name to its contents in string format
func (c *ConfigFetcher) fetch(ctx context.Context, input TerraformOpenSourceParameterParserInput) (map[string]string, error) {
	bucket, key := resolveArtifactPath(input.Artifact.Path)

	configBytes, err := c.s3Downloader.download(ctx, bucket, key)
	if err != nil {
		return map[string]string{},
			ParserAccessDeniedException{Message: fmt.Sprintf(ArtifactFetchAccessDeniedErrorMessage, input.Artifact.Path, err.Error())}
	}

	fileMap, err := UnzipArchive(configBytes)
	if err != nil {
		return fileMap,
			ParserInvalidParameterException{Message: fmt.Sprintf(UnzipFailureErrorMessage, input.Artifact.Path, err.Error())}
	}

	return fileMap, nil
}

func retrieveConfigFetcherCreds(ctx context.Context, launchRoleArn string) (aws.CredentialsProvider, error) {
	if launchRoleArn == "" {
		// use default lambda execution role creds to retrieve configuration templates if launch role is not provided
		log.Print("Launch role is not provided. Using default ServiceCatalogTerraformOSParameterParserRole credentials to fetch artifact.")
		return nil, nil
	}

	log.Printf("Using launch role %s credentials to fetch artifact.", launchRoleArn)
	return retrieveLaunchRoleCreds(ctx, launchRoleArn)
}

// Assumes provided launchRoleArn and return its credentials
func retrieveLaunchRoleCreds(ctx context.Context, launchRoleArn string) (aws.CredentialsProvider, error) {
	// load the default chain's credentials to use for the AssumeRole call itself
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	stsClient := sts.NewFromConfig(cfg)
	return aws.NewCredentialsCache(stscreds.NewAssumeRoleProvider(stsClient, launchRoleArn)), nil
}

// Resolves artifactPath to bucket and key
func resolveArtifactPath(artifactPath string) (string, string) {
	bucket := strings.Split(artifactPath, "/")[2]
	key := strings.SplitN(artifactPath, "/", 4)[3]
	return bucket, key
}
