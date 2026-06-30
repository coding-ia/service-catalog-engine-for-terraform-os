package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
)

type TerraformOpenSourceParameterParserInput struct {
	Artifact      Artifact `json:"artifact"`
	LaunchRoleArn string   `json:"launchRoleArn"`
}

type TerraformOpenSourceParameterParserResponse struct {
	Parameters []*Parameter `json:"parameters"`
}

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, event TerraformOpenSourceParameterParserInput) (TerraformOpenSourceParameterParserResponse, error) {
	if err := ValidateInput(event); err != nil {
		return TerraformOpenSourceParameterParserResponse{}, err
	}

	configFetcher, configFetcherErr := NewConfigFetcher(ctx, event.LaunchRoleArn)
	if configFetcherErr != nil {
		return TerraformOpenSourceParameterParserResponse{}, configFetcherErr
	}

	fileMap, fileMapErr := configFetcher.fetch(ctx, event)
	if fileMapErr != nil {
		return TerraformOpenSourceParameterParserResponse{}, fileMapErr
	}

	parameters, parseParametersErr := ParseParametersFromConfiguration(fileMap)
	return TerraformOpenSourceParameterParserResponse{Parameters: parameters}, parseParametersErr
}
