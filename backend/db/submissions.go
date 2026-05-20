package db

import (
	"context"
	"fmt"

	"github.com/abigailtech/humanloop/backend/models"
)

func CreateSubmission(ctx context.Context, s models.Submission) error {
	_, err := Pool.Exec(ctx, `
		INSERT INTO submissions (id, challenge_id, user_id, video_path, status, latitude, longitude, captured_at, created_at, consent_version, video_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, s.ID, s.ChallengeID, s.UserID, s.VideoPath, "pending", s.Latitude, s.Longitude, s.CapturedAt, s.CreatedAt, s.ConsentVersion, s.VideoHash)
	return err
}

func UpdateSubmissionStatus(ctx context.Context, id, status, hmdfPath string, credits int) error {
	synthetic := status == "synthetic"
	_, err := Pool.Exec(ctx, `
		UPDATE submissions SET status = $1, hmdf_path = $2, credits_earned = $3, synthetic = $4 WHERE id = $5
	`, status, hmdfPath, credits, synthetic, id)
	return err
}

func GetSubmission(ctx context.Context, id string) (models.Submission, error) {
	var s models.Submission
	err := Pool.QueryRow(ctx, `
		SELECT s.id, s.challenge_id, c.title, s.status, s.credits_earned, s.created_at
		FROM submissions s
		LEFT JOIN challenges c ON c.id = s.challenge_id
		WHERE s.id = $1
	`, id).Scan(&s.ID, &s.ChallengeID, &s.ChallengeTitle, &s.Status, &s.CreditsEarned, &s.CreatedAt)
	return s, err
}

func LogCreditTransaction(ctx context.Context, userID, submissionID, reason string, amount int) error {
	_, err := Pool.Exec(ctx, `
		INSERT INTO credit_transactions (user_id, submission_id, amount, reason)
		VALUES ($1, $2, $3, $4)
	`, userID, submissionID, amount, reason)
	return err
}

func SetSubmissionQuality(ctx context.Context, id string, qualityScore float64, extractorVersion string) error {
	_, err := Pool.Exec(ctx, `
		UPDATE submissions SET quality_score = $1, extractor_version = $2 WHERE id = $3
	`, qualityScore, extractorVersion, id)
	return err
}

func VideoHashExists(ctx context.Context, hash string) (bool, error) {
	var exists bool
	err := Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM submissions WHERE video_hash = $1 AND video_hash != '')`, hash).Scan(&exists)
	return exists, err
}

func CountSubmissionsToday(ctx context.Context, userID string) (int, error) {
	var count int
	err := Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM submissions WHERE user_id = $1 AND created_at > NOW() - INTERVAL '24 hours'
	`, userID).Scan(&count)
	return count, err
}

func ApproveSubmission(ctx context.Context, id string, approved bool) error {
	_, err := Pool.Exec(ctx, `UPDATE submissions SET approved = $1 WHERE id = $2`, approved, id)
	return err
}

func TagSubmission(ctx context.Context, id string, tags []string) error {
	_, err := Pool.Exec(ctx, `UPDATE submissions SET tags = $1 WHERE id = $2`, tags, id)
	return err
}

func GetAdminSubmissions(ctx context.Context, status, robot string, approved *bool, limit, offset int) ([]models.Submission, error) {
	args := []any{}
	where := "WHERE 1=1"
	i := 1
	if status != "" {
		where += fmt.Sprintf(" AND s.status = $%d", i)
		args = append(args, status)
		i++
	}
	if robot != "" {
		where += fmt.Sprintf(" AND s.robot = $%d", i)
		args = append(args, robot)
		i++
	}
	if approved != nil {
		where += fmt.Sprintf(" AND s.approved = $%d", i)
		args = append(args, *approved)
		i++
	}
	args = append(args, limit, offset)
	rows, err := Pool.Query(ctx, fmt.Sprintf(`
		SELECT s.id, s.challenge_id, c.title, s.status, s.credits_earned, s.quality_score,
		       s.extractor_version, s.approved, s.tags, s.created_at
		FROM submissions s
		LEFT JOIN challenges c ON c.id = s.challenge_id
		%s ORDER BY s.created_at DESC LIMIT $%d OFFSET $%d
	`, where, i, i+1), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Submission
	for rows.Next() {
		var s models.Submission
		if err := rows.Scan(&s.ID, &s.ChallengeID, &s.ChallengeTitle, &s.Status, &s.CreditsEarned,
			&s.QualityScore, &s.ExtractorVersion, &s.Approved, &s.Tags, &s.CreatedAt); err != nil {
			continue
		}
		list = append(list, s)
	}
	return list, nil
}

func GetSubmissionsByUser(ctx context.Context, userID string, limit, offset int) ([]models.Submission, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := Pool.Query(ctx, `
		SELECT s.id, s.challenge_id, c.title, s.status, s.credits_earned, s.quality_score, s.approved, s.created_at
		FROM submissions s
		LEFT JOIN challenges c ON c.id = s.challenge_id
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Submission
	for rows.Next() {
		var s models.Submission
		if err := rows.Scan(&s.ID, &s.ChallengeID, &s.ChallengeTitle, &s.Status, &s.CreditsEarned, &s.QualityScore, &s.Approved, &s.CreatedAt); err != nil {
			continue
		}
		list = append(list, s)
	}
	return list, nil
}
