package db

import (
	"context"

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
		SELECT id, email, name, credits, submissions FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.Email, &u.Name, &u.Credits, &u.Submissions)
	return u, err
}

func AddCredits(ctx context.Context, userID string, amount int) error {
	_, err := Pool.Exec(ctx, `
		UPDATE users SET credits = credits + $1, submissions = submissions + 1 WHERE id = $2
	`, amount, userID)
	return err
}

func GetLeaderboard(ctx context.Context) ([]models.User, error) {
	rows, err := Pool.Query(ctx, `
		SELECT id, name, credits, submissions FROM users ORDER BY credits DESC LIMIT 100
	`)
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
