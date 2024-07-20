package config

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
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
	GCPCredentialsPath string
	VertexAILocation   string
	Port               string
}

type ProviderConfig struct {
	Models   []string `toml:"models"`
	Prompt   string   `toml:"prompt"`
	Endpoint string   `toml:"endpoint"`
}

type DefaultConfig struct {
	Providers []string `toml:"providers"`
}

func Load() (*Config, error) {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "local" // デフォルトはローカル環境
	}

	var configData []byte
	var err error

	if environment == "local" {
		configData, err = loadLocalConfig()
		if err != nil {
			// ローカル設定の読み込みに失敗した場合、Cloud Storageから読み込む
			configData, err = loadCloudStorageConfig()
		}
	} else {
		configData, err = loadCloudStorageConfig()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	var cfg Config
	if _, err := toml.Decode(string(configData), &cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %v", err)
	}

	// 環境変数から追加の設定を読み込む
	cfg.AWSAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	cfg.AWSSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	cfg.AWSRegion = os.Getenv("AWS_REGION")
	cfg.OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")
	cfg.GCPProjectID = os.Getenv("GCP_PROJECT_ID")
	cfg.BigQueryDatasetID = os.Getenv("BIGQUERY_DATASET_ID")
	cfg.BigQueryTableID = os.Getenv("BIGQUERY_TABLE_ID")
	cfg.GCPCredentialsPath = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	cfg.VertexAILocation = os.Getenv("VERTEX_AI_LOCATION")
	cfg.Port = os.Getenv("PORT")

	return &cfg, nil
}

func loadLocalConfig() ([]byte, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.toml" // デフォルトのパス
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read local config file: %v", err)
	}

	return data, nil
}

func loadCloudStorageConfig() ([]byte, error) {
	bucketName := os.Getenv("CONFIG_BUCKET")
	objectName := os.Getenv("CONFIG_OBJECT")

	if bucketName == "" || objectName == "" {
		return nil, fmt.Errorf("CONFIG_BUCKET and CONFIG_OBJECT must be set")
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)
	obj := bucket.Object(objectName)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create object reader: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %v", err)
	}

	return data, nil
}

func (c *Config) GetDefaultProviders() []string {
	return c.Default.Providers
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
