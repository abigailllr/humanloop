package db

import (
	"context"
	"time"
)

func CreateRefreshToken(ctx context.Context, id, userID, tokenHash string) error {
	_, err := Pool.Exec(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
	`, id, userID, tokenHash, time.Now().Add(30*24*time.Hour))
	return err
}

func GetRefreshToken(ctx context.Context, tokenHash string) (userID, id string, err error) {
	err = Pool.QueryRow(ctx, `
		SELECT user_id, id FROM refresh_tokens
		WHERE token_hash = $1 AND expires_at > NOW()
	`, tokenHash).Scan(&userID, &id)
	return
}

func DeleteRefreshToken(ctx context.Context, id string) error {
	_, err := Pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE id = $1`, id)
	return err
}

func DeleteRefreshTokenByHash(ctx context.Context, hash string) error {
	_, err := Pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE token_hash = $1`, hash)
	return err
}

func PurgeExpiredRefreshTokens(ctx context.Context) error {
	_, err := Pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE expires_at <= NOW()`)
	return err
}
