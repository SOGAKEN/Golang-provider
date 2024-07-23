package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"provider/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type AWSBedrockProvider struct {
	client           *bedrockruntime.Client
	models           []string
	anthropicVersion string
}

func NewAWSBedrockProvider(cfg *config.Config) (Provider, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.AWSRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWSAccessKeyID,
			cfg.AWSSecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := bedrockruntime.NewFromConfig(awsCfg)

	models, err := cfg.GetProviderModels("aws_bedrock")
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS Bedrock models: %w", err)
	}

	return &AWSBedrockProvider{
		client:           client,
		models:           models,
		anthropicVersion: cfg.AWSAnthropicVersion,
	}, nil
}

func (p *AWSBedrockProvider) Generate(prompt string, model string) (string, error) {
	requestBody := map[string]interface{}{
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":  300,
		"temperature": 0.7,
		"top_p":       0.9,
	}

	// anthropic_versionが設定されている場合のみ追加
	if p.anthropicVersion != "" {
		requestBody["anthropic_version"] = p.anthropicVersion
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	input := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(model),
		Body:        jsonBody,
		ContentType: aws.String("application/json"),
	}

	output, err := p.client.InvokeModel(context.TODO(), input)
	if err != nil {
		return "", fmt.Errorf("Bedrock API call failed: %w", err)
	}

	var response struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(output.Body, &response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Content) == 0 || len(response.Content[0].Text) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return response.Content[0].Text, nil
}

func (p *AWSBedrockProvider) GetModels() []string {
	return p.models
}

