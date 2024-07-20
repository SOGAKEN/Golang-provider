package providers

import (
	"fmt"
	"provider/internal/config"
)

type Provider interface {
	Generate(prompt string, model string) (string, error)
	GetModels() []string
}

func GetProvider(cfg *config.Config, providerName string) (Provider, error) {
	switch providerName {
	case "aws_bedrock":
		return NewAWSBedrockProvider(cfg)
	case "openai":
		return NewOpenAIProvider(cfg)
	case "vertexai_gemini":
		return NewVertexAIGeminiProvider(cfg)
	default:
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}
}
