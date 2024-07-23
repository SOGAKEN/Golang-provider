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
	// .env ファイルを読み込む
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or unable to load")
	}

	// 初期設定の読み込み（エラーチェックのみ）
	initialCfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load initial config: %v", err)
	}

	// 環境変数が設定されているか確認、未設定の場合はデフォルト値を使用
	port := initialCfg.Port
	if port == "" {
		port = "8080" // デフォルトポートを8080に設定
		log.Printf("Defaulting to port %s", port)
	}

	bqClient, err := storage.NewBigQueryClient(initialCfg.GCPProjectID, initialCfg.BigQueryDatasetID, initialCfg.BigQueryTableID)
	if err != nil {
		log.Fatalf("Failed to create BigQuery client: %v", err)
	}
	defer bqClient.Close()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	api.SetupRoutes(r, bqClient)

	log.Printf("Server is starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

