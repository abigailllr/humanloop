package db

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/abigailtech/humanloop/backend/models"
)

func UpsertUser(ctx context.Context, u models.User) error {
	_, err := Pool.Exec(ctx, `
		INSERT INTO users (id, email, name)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET email = EXCLUDED.email, name = EXCLUDED.name
	`, u.ID, u.Email, u.Name)
	return err
}

func GetUser(ctx context.Context, id string) (models.User, error) {
	var u models.User
	err := Pool.QueryRow(ctx, `
		SELECT id, email, name, credits, submissions, COALESCE(referral_code,'') FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.Email, &u.Name, &u.Credits, &u.Submissions, &u.ReferralCode)
	return u, err
}

func AddCredits(ctx context.Context, userID string, amount int) error {
	_, err := Pool.Exec(ctx, `
		UPDATE users SET credits = credits + $1, submissions = submissions + 1 WHERE id = $2
	`, amount, userID)
	return err
}

func GetLeaderboard(ctx context.Context, period string) ([]models.User, error) {
	const base = `SELECT id, name, credits, submissions FROM users`
	const order = ` ORDER BY credits DESC LIMIT 100`
	const recent = ` WHERE id IN (SELECT DISTINCT user_id FROM submissions WHERE status='done' AND created_at > NOW() - ($1 || ' days')::interval)`

	var query string
	var rows pgx.Rows
	var err error
	switch period {
	case "week":
		query = base + recent + order
		rows, err = Pool.Query(ctx, query, "7")
	case "month":
		query = base + recent + order
		rows, err = Pool.Query(ctx, query, "30")
	default:
		query = base + order
		rows, err = Pool.Query(ctx, query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Credits, &u.Submissions); err != nil {
			continue
		}
		list = append(list, u)
	}
	return list, nil
}

func DeleteUser(ctx context.Context, userID string) error {
	for _, q := range []string{
		`DELETE FROM refresh_tokens WHERE user_id = $1`,
		`DELETE FROM credit_transactions WHERE user_id = $1`,
		`DELETE FROM referrals WHERE referrer_id = $1 OR referee_id = $1`,
		`DELETE FROM submissions WHERE user_id = $1`,
	} {
		if _, err := Pool.Exec(ctx, q, userID); err != nil {
			return err
		}
	}
	_, err := Pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	return err
}
