package providers

import (
	"context"
	"fmt"

	"provider/internal/config"

	"github.com/sashabaranov/go-openai"
)

type OpenAIProvider struct {
	client *openai.Client
	models []string
}

func NewOpenAIProvider(cfg *config.Config) (Provider, error) {
	endpoint, err := cfg.GetProviderEndpoint("openai")
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenAI endpoint: %w", err)
	}

	clientConfig := openai.DefaultConfig(cfg.OpenAIAPIKey)
	clientConfig.BaseURL = endpoint

	client := openai.NewClientWithConfig(clientConfig)

	models, err := cfg.GetProviderModels("openai")
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenAI models: %w", err)
	}

	return &OpenAIProvider{
		client: client,
		models: models,
	}, nil
}

func (p *OpenAIProvider) Generate(prompt string, model string) (string, error) {
	resp, err := p.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

func (p *OpenAIProvider) GetModels() []string {
	return p.models
}

