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
	GCPProjectID       string
	BigQueryDatasetID  string
	BigQueryTableID    string
}

type ProviderConfig struct {
	Models []string `toml:"models"`
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
	cfg.GCPProjectID = os.Getenv("GCP_PROJECT_ID")
	cfg.BigQueryDatasetID = os.Getenv("BIGQUERY_DATASET_ID")
	cfg.BigQueryTableID = os.Getenv("BIGQUERY_TABLE_ID")

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
