package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Providers map[string]ProviderConfig `toml:"providers"`
	Default   DefaultConfig             `toml:"default"`

	// 環境変数から読み込む設定
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
	OpenAIAPIKey       string
	GCPProjectID       string
	BigQueryDatasetID  string
	BigQueryTableID    string
	Port               string
}

type ProviderConfig struct {
	Models   []string `toml:"models"`
	Prompt   string   `toml:"prompt"`
	Endpoint string   `toml:"endpoint"`
}

type DefaultConfig struct {
	Provider string `toml:"provider"`
}

func Load() (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile("config.toml", &cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	// 環境変数から追加の設定を読み込む
	cfg.AWSAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	cfg.AWSSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	cfg.AWSRegion = os.Getenv("AWS_REGION")
	cfg.OpenAIAPIKey = os.Getenv("OPENAI_API_KEY") // 追加
	cfg.GCPProjectID = os.Getenv("GCP_PROJECT_ID")
	cfg.BigQueryDatasetID = os.Getenv("BIGQUERY_DATASET_ID")
	cfg.BigQueryTableID = os.Getenv("BIGQUERY_TABLE_ID")
	cfg.Port = os.Getenv("PORT") // 追加

	// ポートが設定されていない場合のデフォルト値
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return &cfg, nil
}

func (c *Config) GetDefaultProvider() string {
	return c.Default.Provider
}

func (c *Config) GetProviderModels(provider string) ([]string, error) {
	if providerConfig, ok := c.Providers[provider]; ok {
		return providerConfig.Models, nil
	}
	return nil, fmt.Errorf("provider not found: %s", provider)
}

func (c *Config) GetProviderPrompt(provider string) (string, error) {
	if providerConfig, ok := c.Providers[provider]; ok {
		return providerConfig.Prompt, nil
	}
	return "", fmt.Errorf("provider not found: %s", provider)
}

func (c *Config) GetProviderEndpoint(provider string) (string, error) {
	if providerConfig, ok := c.Providers[provider]; ok {
		return providerConfig.Endpoint, nil
	}
	return "", fmt.Errorf("provider not found: %s", provider)
}
