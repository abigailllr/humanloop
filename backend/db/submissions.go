package db

import (
	"context"

	"github.com/abigailtech/humanloop/backend/models"
)

func CreateSubmission(ctx context.Context, s models.Submission) error {
	_, err := Pool.Exec(ctx, `
		INSERT INTO submissions (id, challenge_id, user_id, video_path, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, s.ID, s.ChallengeID, s.UserID, s.VideoPath, "pending", s.CreatedAt)
	return err
}

func UpdateSubmissionStatus(ctx context.Context, id, status, hmdfPath string) error {
	_, err := Pool.Exec(ctx, `
		UPDATE submissions SET status = $1, hmdf_path = $2 WHERE id = $3
	`, status, hmdfPath, id)
	return err
}

func GetSubmissionsByUser(ctx context.Context, userID string) ([]models.Submission, error) {
	rows, err := Pool.Query(ctx, `
		SELECT id, challenge_id, user_id, video_path, status, created_at
		FROM submissions WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Submission
	for rows.Next() {
		var s models.Submission
		if err := rows.Scan(&s.ID, &s.ChallengeID, &s.UserID, &s.VideoPath, &s.Status, &s.CreatedAt); err != nil {
			continue
		}
		list = append(list, s)
	}
	return list, nil
}
