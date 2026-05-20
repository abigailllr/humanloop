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

func GetChallengeStats(ctx context.Context, challengeID string) (map[string]any, error) {
	var total, approved, synthetic int
	var avgQ, maxQ float64
	err := Pool.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE approved AND status='done'),
		       COUNT(*) FILTER (WHERE synthetic),
		       COALESCE(AVG(quality_score) FILTER (WHERE status='done'), 0),
		       COALESCE(MAX(quality_score), 0)
		FROM submissions WHERE challenge_id = $1
	`, challengeID).Scan(&total, &approved, &synthetic, &avgQ, &maxQ)
	if err != nil {
		return nil, err
	}

	rows, err := Pool.Query(ctx, `
		SELECT robot, COUNT(*) FROM submissions WHERE challenge_id = $1 AND status = 'done' GROUP BY robot ORDER BY COUNT(*) DESC
	`, challengeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	robots := map[string]int{}
	for rows.Next() {
		var r string
		var c int
		if rows.Scan(&r, &c) == nil {
			robots[r] = c
		}
	}

	return map[string]any{
		"challenge_id":        challengeID,
		"total_submissions":   total,
		"approved_done":       approved,
		"synthetic_rejected":  synthetic,
		"quality_avg":         avgQ,
		"quality_max":         maxQ,
		"submissions_by_robot": robots,
	}, nil
}

func GetChallengeLeaderboard(ctx context.Context, challengeID string, limit int) ([]models.User, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := Pool.Query(ctx, `
		SELECT u.id, u.name, COUNT(s.id) as sub_count, COALESCE(SUM(s.credits_earned), 0) as earned
		FROM submissions s
		JOIN users u ON u.id = s.user_id
		WHERE s.challenge_id = $1 AND s.status = 'done'
		GROUP BY u.id, u.name ORDER BY earned DESC, sub_count DESC LIMIT $2
	`, challengeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Submissions, &u.Credits); err != nil {
			continue
		}
		list = append(list, u)
	}
	return list, nil
}
