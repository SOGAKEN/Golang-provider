package providers

import (
	"context"
	"fmt"

	"provider/internal/config"

	"cloud.google.com/go/vertexai/genai"
	"google.golang.org/api/option"
)

type VertexAIGeminiProvider struct {
	client    *genai.Client
	models    []string
	projectID string
	location  string
}

func NewVertexAIGeminiProvider(cfg *config.Config) (Provider, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, cfg.GCPProjectID, cfg.VertexAILocation, option.WithCredentialsFile(cfg.GCPCredentialsPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create Vertex AI client: %v", err)
	}

	models, err := cfg.GetProviderModels("vertexai_gemini")
	if err != nil {
		return nil, fmt.Errorf("failed to get Vertex AI Gemini models: %w", err)
	}

	return &VertexAIGeminiProvider{
		client:    client,
		models:    models,
		projectID: cfg.GCPProjectID,
		location:  cfg.VertexAILocation,
	}, nil
}

func (p *VertexAIGeminiProvider) Generate(prompt string, model string) (string, error) {
	ctx := context.Background()

	geminiModel := p.client.GenerativeModel(model)
	geminiModel.SetTemperature(0.2)
	geminiModel.SetTopK(40)
	geminiModel.SetTopP(0.95)
	geminiModel.SetMaxOutputTokens(1024)

	resp, err := geminiModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	generatedText, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", fmt.Errorf("unexpected content type")
	}

	return string(generatedText), nil
}

func (p *VertexAIGeminiProvider) GetModels() []string {
	return p.models
}

