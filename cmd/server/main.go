package main

import (
	"log"

	"provider/internal/api"
	"provider/internal/config"
	"provider/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
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

	// サーバー起動時にマイグレーションを実行
	if err := bqClient.MigrateTable(); err != nil {
		log.Fatalf("Failed to migrate table: %v", err)
	}

	//	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	api.SetupRoutes(r, cfg, bqClient)

	r.Run(":8080")
}
