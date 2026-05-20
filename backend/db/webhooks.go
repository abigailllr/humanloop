package db

import (
	"context"

	"github.com/abigailtech/humanloop/backend/models"
)

func CreateWebhook(ctx context.Context, w models.Webhook) error {
	_, err := Pool.Exec(ctx, `
		INSERT INTO webhooks (id, dataset_id, url, secret_hash, active)
		VALUES ($1, NULLIF($2,''), $3, $4, TRUE)
	`, w.ID, w.DatasetID, w.URL, w.SecretHash)
	return err
}

func GetActiveWebhooksForDataset(ctx context.Context, datasetID string) ([]models.Webhook, error) {
	rows, err := Pool.Query(ctx, `
		SELECT id, COALESCE(dataset_id,''), url, secret_hash, active, created_at
		FROM webhooks WHERE (dataset_id = $1 OR dataset_id IS NULL) AND active = TRUE
	`, datasetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWebhooks(rows)
}

func ListWebhooks(ctx context.Context) ([]models.Webhook, error) {
	rows, err := Pool.Query(ctx, `
		SELECT id, COALESCE(dataset_id,''), url, secret_hash, active, created_at
		FROM webhooks ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWebhooks(rows)
}

func DeleteWebhook(ctx context.Context, id string) error {
	_, err := Pool.Exec(ctx, `DELETE FROM webhooks WHERE id = $1`, id)
	return err
}

type webhookRows interface {
	Next() bool
	Scan(dest ...any) error
}

func scanWebhooks(rows webhookRows) ([]models.Webhook, error) {
	var list []models.Webhook
	for rows.Next() {
		var w models.Webhook
		if err := rows.Scan(&w.ID, &w.DatasetID, &w.URL, &w.SecretHash, &w.Active, &w.CreatedAt); err != nil {
			continue
		}
		list = append(list, w)
	}
	return list, nil
}
