package db

import (
	"context"

	"github.com/abigailtech/humanloop/backend/models"
)

func CreateSubmission(ctx context.Context, s models.Submission) error {
	_, err := Pool.Exec(ctx, `
		INSERT INTO submissions (id, challenge_id, user_id, video_path, status, latitude, longitude, captured_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, s.ID, s.ChallengeID, s.UserID, s.VideoPath, "pending", s.Latitude, s.Longitude, s.CapturedAt, s.CreatedAt)
	return err
}

func UpdateSubmissionStatus(ctx context.Context, id, status, hmdfPath string, credits int) error {
	_, err := Pool.Exec(ctx, `
		UPDATE submissions SET status = $1, hmdf_path = $2, credits_earned = $3 WHERE id = $4
	`, status, hmdfPath, credits, id)
	return err
}

func GetSubmissionsByUser(ctx context.Context, userID string) ([]models.Submission, error) {
	rows, err := Pool.Query(ctx, `
		SELECT s.id, s.challenge_id, c.title, s.status, s.credits_earned, s.created_at
		FROM submissions s
		LEFT JOIN challenges c ON c.id = s.challenge_id
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Submission
	for rows.Next() {
		var s models.Submission
		if err := rows.Scan(&s.ID, &s.ChallengeID, &s.ChallengeTitle, &s.Status, &s.CreditsEarned, &s.CreatedAt); err != nil {
			continue
		}
		list = append(list, s)
	}
	return list, nil
}
