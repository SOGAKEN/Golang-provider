package config

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"cloud.google.com/go/storage"
	"github.com/BurntSushi/toml"
)

type Config struct {
	Providers map[string]ProviderConfig `toml:"providers"`
	Default   DefaultConfig             `toml:"default"`

	// 環境変数から読み込む設定
	AWSAccessKeyID      string
	AWSSecretAccessKey  string
	AWSRegion           string
	AWSAnthropicVersion string
	OpenAIAPIKey        string
	GCPProjectID        string
	BigQueryDatasetID   string
	BigQueryTableID     string
	GCPCredentialsPath  string
	VertexAILocation    string
	Port                string
}

type ProviderConfig struct {
	Models   []string `toml:"models"`
	Prompt   string   `toml:"prompt"`
	Endpoint string   `toml:"endpoint"`
}

type DefaultConfig struct {
	Providers []string `toml:"providers"`
}

var (
	globalConfig     *Config
	globalConfigOnce sync.Once
)

func Load() (*Config, error) {
	var err error
	globalConfigOnce.Do(func() {
		globalConfig, err = loadConfig()
	})
	if err != nil {
		return nil, err
	}
	return globalConfig, nil
}

func loadConfig() (*Config, error) {
	var cfg Config

	// 環境変数から設定を読み込む
	cfg.AWSAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	cfg.AWSSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	cfg.AWSRegion = os.Getenv("AWS_REGION")
	cfg.AWSAnthropicVersion = os.Getenv("AWS_ANTHROPIC_VERSION")
	cfg.OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")
	cfg.GCPProjectID = os.Getenv("GCP_PROJECT_ID")
	cfg.BigQueryDatasetID = os.Getenv("BIGQUERY_DATASET_ID")
	cfg.BigQueryTableID = os.Getenv("BIGQUERY_TABLE_ID")
	cfg.GCPCredentialsPath = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	cfg.VertexAILocation = os.Getenv("VERTEX_AI_LOCATION")
	cfg.Port = os.Getenv("PORT")

	return &cfg, nil
}

func LoadTOMLConfig() (*Config, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	// Cloud Storageからconfig.tomlを読み込む
	tomlData, err := loadFromCloudStorage()
	if err != nil {
		log.Printf("Failed to load config from Cloud Storage: %v. Falling back to local config.", err)
		tomlData, err = loadFromLocalFile()
		if err != nil {
			return nil, fmt.Errorf("failed to load config from both Cloud Storage and local file: %v", err)
		}
	}

	// TOMLデータをデコード
	if _, err := toml.Decode(string(tomlData), cfg); err != nil {
		return nil, fmt.Errorf("failed to decode TOML config: %v", err)
	}

	return cfg, nil
}

func loadFromCloudStorage() ([]byte, error) {
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

	return io.ReadAll(reader)
}

func loadFromLocalFile() ([]byte, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.toml" // デフォルトのパス
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}

	return os.ReadFile(absPath)
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
