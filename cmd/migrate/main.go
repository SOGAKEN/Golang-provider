package main

import (
	"log"

	"provider/internal/config"
	"provider/internal/storage"

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

	if err := bqClient.MigrateTable(); err != nil {
		log.Fatalf("Failed to migrate table: %v", err)
	}

	log.Println("Migration completed successfully.")
}
