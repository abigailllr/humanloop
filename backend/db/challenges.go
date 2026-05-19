package db

import (
	"context"

	"github.com/abigailtech/humanloop/backend/models"
)

var DefaultChallenges = []models.Challenge{
	{ID: "c1", Title: "Pick & Place", Description: "Pick up any object from a table and place it into a box.", Difficulty: "Easy"},
	{ID: "c2", Title: "Fold It", Description: "Fold a piece of cloth or paper in half.", Difficulty: "Easy"},
	{ID: "c3", Title: "Sort & Stack", Description: "Sort 5 objects by size from smallest to largest.", Difficulty: "Medium"},
}

func SeedChallenges(ctx context.Context) error {
	for _, c := range DefaultChallenges {
		_, err := Pool.Exec(ctx, `
			INSERT INTO challenges (id, title, description, difficulty)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO NOTHING
		`, c.ID, c.Title, c.Description, c.Difficulty)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetChallenges(ctx context.Context) ([]models.Challenge, error) {
	rows, err := Pool.Query(ctx, `SELECT id, title, description, difficulty, submissions FROM challenges ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Challenge
	for rows.Next() {
		var c models.Challenge
		if err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.Difficulty, &c.Submissions); err != nil {
			continue
		}
		list = append(list, c)
	}
	return list, nil
}

func IncrementChallengeSubmissions(ctx context.Context, challengeID string) error {
	_, err := Pool.Exec(ctx, `UPDATE challenges SET submissions = submissions + 1 WHERE id = $1`, challengeID)
	return err
}

func CreateChallenge(ctx context.Context, c models.Challenge) error {
	_, err := Pool.Exec(ctx, `
		INSERT INTO challenges (id, title, description, difficulty)
		VALUES ($1, $2, $3, $4)
	`, c.ID, c.Title, c.Description, c.Difficulty)
	return err
}

func UpdateChallenge(ctx context.Context, c models.Challenge) error {
	_, err := Pool.Exec(ctx, `
		UPDATE challenges SET title=$2, description=$3, difficulty=$4 WHERE id=$1
	`, c.ID, c.Title, c.Description, c.Difficulty)
	return err
}

func DeleteChallenge(ctx context.Context, id string) error {
	_, err := Pool.Exec(ctx, `DELETE FROM challenges WHERE id=$1`, id)
	return err
}
