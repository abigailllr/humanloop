package db

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/abigailtech/humanloop/backend/models"
)

func HashKey(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func CreateBuyerKey(ctx context.Context, k models.BuyerKey, rawKey string) error {
	_, err := Pool.Exec(ctx, `
		INSERT INTO buyer_keys (id, key_hash, label, dataset_id) VALUES ($1, $2, $3, NULLIF($4, ''))
	`, k.ID, HashKey(rawKey), k.Label, k.DatasetID)
	return err
}

func GetBuyerKeyByHash(ctx context.Context, hash string) (models.BuyerKey, error) {
	var k models.BuyerKey
	var datasetID *string
	err := Pool.QueryRow(ctx, `SELECT id, label, dataset_id, created_at FROM buyer_keys WHERE key_hash = $1`, hash).
		Scan(&k.ID, &k.Label, &datasetID, &k.CreatedAt)
	if datasetID != nil {
		k.DatasetID = *datasetID
	}
	return k, err
}

func ListBuyerKeys(ctx context.Context) ([]models.BuyerKey, error) {
	rows, err := Pool.Query(ctx, `SELECT id, label, dataset_id, created_at FROM buyer_keys ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.BuyerKey
	for rows.Next() {
		var k models.BuyerKey
		var datasetID *string
		if err := rows.Scan(&k.ID, &k.Label, &datasetID, &k.CreatedAt); err != nil {
			continue
		}
		if datasetID != nil {
			k.DatasetID = *datasetID
		}
		list = append(list, k)
	}
	return list, nil
}

func DeleteBuyerKey(ctx context.Context, id string) error {
	_, err := Pool.Exec(ctx, `DELETE FROM buyer_keys WHERE id = $1`, id)
	return err
}
