package storage

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/googleapi"
)

func (bq *BigQueryClient) MigrateTable() error {
	ctx := context.Background()
	schema := bigquery.Schema{
		{Name: "Timestamp", Type: bigquery.TimestampFieldType},
		{Name: "ResponseTime", Type: bigquery.FloatFieldType},
		{Name: "RequestBody", Type: bigquery.StringFieldType},
		{Name: "ResponseBody", Type: bigquery.StringFieldType},
		{Name: "Model", Type: bigquery.StringFieldType},
		{Name: "Provider", Type: bigquery.StringFieldType},
	}

	dataset := bq.client.Dataset(bq.datasetID)
	tableRef := dataset.Table(bq.tableID)

	metadata, err := tableRef.Metadata(ctx)
	if err != nil {
		if e, ok := err.(*googleapi.Error); ok && e.Code == 404 {
			// テーブルが存在しない場合は作成
			return bq.createTable(ctx, schema)
		}
		return fmt.Errorf("failed to get table metadata: %v", err)
	}

	// テーブルが存在する場合はスキーマを更新
	return bq.updateTableSchema(ctx, metadata, schema)
}

func (bq *BigQueryClient) createTable(ctx context.Context, schema bigquery.Schema) error {
	tableRef := bq.client.Dataset(bq.datasetID).Table(bq.tableID)
	if err := tableRef.Create(ctx, &bigquery.TableMetadata{Schema: schema}); err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}
	fmt.Printf("Table %s created.\n", bq.tableID)
	return nil
}

func (bq *BigQueryClient) updateTableSchema(ctx context.Context, metadata *bigquery.TableMetadata, newSchema bigquery.Schema) error {
	tableRef := bq.client.Dataset(bq.datasetID).Table(bq.tableID)
	update := bigquery.TableMetadataToUpdate{
		Schema: newSchema,
	}
	if _, err := tableRef.Update(ctx, update, metadata.ETag); err != nil {
		return fmt.Errorf("failed to update table schema: %v", err)
	}
	fmt.Printf("Table %s schema updated.\n", bq.tableID)
	return nil
}
