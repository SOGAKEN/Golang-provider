package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/option"
)

type GenerationLog struct {
	Timestamp    time.Time
	ResponseTime float64
	RequestBody  string
	ResponseBody string
	Model        string
	Provider     string
}

type BigQueryClient struct {
	projectID string
	datasetID string
	tableID   string
	client    *bigquery.Client
}

func NewBigQueryClient(projectID, datasetID, tableID string) (*BigQueryClient, error) {
	ctx := context.Background()

	// 環境変数からサービスアカウントキーの内容を取得
	saKey := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if saKey == "" {
		return nil, fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS environment variable is not set")
	}

	// サービスアカウントキーを使用してクライアントを作成
	client, err := bigquery.NewClient(ctx, projectID, option.WithCredentialsJSON([]byte(saKey)))
	if err != nil {
		return nil, fmt.Errorf("bigquery.NewClient: %v", err)
	}

	return &BigQueryClient{
		projectID: projectID,
		datasetID: datasetID,
		tableID:   tableID,
		client:    client,
	}, nil
}

func (bq *BigQueryClient) InsertGenerationLog(logEntry *GenerationLog) error {
	ctx := context.Background()
	inserter := bq.client.Dataset(bq.datasetID).Table(bq.tableID).Inserter()

	log.Printf("Inserting log into BigQuery. Dataset: %s, Table: %s", bq.datasetID, bq.tableID)

	if err := inserter.Put(ctx, logEntry); err != nil {
		log.Printf("Error inserting data into BigQuery: %v", err)
		return fmt.Errorf("failed to insert data into BigQuery: %v", err)
	}

	log.Println("Log inserted successfully into BigQuery")
	return nil
}

func (bq *BigQueryClient) Close() error {
	return bq.client.Close()
}
