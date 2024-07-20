package main

import (
	"log"
	"os"

	"provider/internal/api"
	"provider/internal/config"
	"provider/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// .env ファイルを読み込む
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or unable to load")
	}

	// 環境変数が設定されているか確認
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		log.Fatal("GOOGLE_APPLICATION_CREDENTIALS environment variable is not set")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	bqClient, err := storage.NewBigQueryClient(cfg.GCPProjectID, cfg.BigQueryDatasetID, cfg.BigQueryTableID)
	if err != nil {
		log.Fatalf("Failed to create BigQuery client: %v", err)
	}
	defer bqClient.Close()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	api.SetupRoutes(r, cfg, bqClient)

	log.Printf("Server is starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
