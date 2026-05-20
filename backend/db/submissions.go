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
	var hmdfPath *string
	err := Pool.QueryRow(ctx, `
		SELECT s.id, s.challenge_id, c.title, s.status, s.credits_earned, s.hmdf_path, s.created_at
		FROM submissions s
		LEFT JOIN challenges c ON c.id = s.challenge_id
		WHERE s.id = $1
	`, id).Scan(&s.ID, &s.ChallengeID, &s.ChallengeTitle, &s.Status, &s.CreditsEarned, &hmdfPath, &s.CreatedAt)
	if hmdfPath != nil {
		s.HmdfPath = *hmdfPath
	}
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

func GetCreditHistory(ctx context.Context, userID string, limit, offset int) ([]map[string]any, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := Pool.Query(ctx, `
		SELECT id, submission_id, amount, reason, created_at
		FROM credit_transactions WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []map[string]any
	for rows.Next() {
		var id int64
		var submissionID *string
		var amount int
		var reason, createdAt string
		if err := rows.Scan(&id, &submissionID, &amount, &reason, &createdAt); err != nil {
			continue
		}
		entry := map[string]any{"id": id, "amount": amount, "reason": reason, "created_at": createdAt}
		if submissionID != nil {
			entry["submission_id"] = *submissionID
		}
		list = append(list, entry)
	}
	return list, nil
}

func GetUserStats(ctx context.Context, userID string) (map[string]any, error) {
	var total, done, failed, synthetic int
	var avgQ float64
	err := Pool.QueryRow(ctx, `
		SELECT COUNT(*),
		       COUNT(*) FILTER (WHERE status='done'),
		       COUNT(*) FILTER (WHERE status='failed'),
		       COUNT(*) FILTER (WHERE synthetic),
		       COALESCE(AVG(quality_score) FILTER (WHERE status='done'), 0)
		FROM submissions WHERE user_id = $1
	`, userID).Scan(&total, &done, &failed, &synthetic, &avgQ)
	if err != nil {
		return nil, err
	}
	rows, err := Pool.Query(ctx, `
		SELECT robot, COUNT(*) FROM submissions WHERE user_id = $1 AND status = 'done' GROUP BY robot
	`, userID)
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
		"total_submissions":  total,
		"done":               done,
		"failed":             failed,
		"synthetic_rejected": synthetic,
		"quality_avg":        avgQ,
		"robots_used":        robots,
	}, nil
}

func GetSubmissionHeatmap(ctx context.Context) ([]map[string]any, error) {
	rows, err := Pool.Query(ctx, `
		SELECT ROUND(latitude::numeric, 1) AS lat, ROUND(longitude::numeric, 1) AS lng, COUNT(*)
		FROM submissions
		WHERE status = 'done' AND latitude IS NOT NULL AND longitude IS NOT NULL
		  AND latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180
		GROUP BY lat, lng
		ORDER BY COUNT(*) DESC
		LIMIT 5000
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []map[string]any
	for rows.Next() {
		var lat, lng float64
		var count int
		if err := rows.Scan(&lat, &lng, &count); err != nil {
			continue
		}
		list = append(list, map[string]any{"lat": lat, "lng": lng, "count": count})
	}
	return list, nil
}

func IncrementRetryCount(ctx context.Context, id string) (int, error) {
	var count int
	err := Pool.QueryRow(ctx, `
		UPDATE submissions SET retry_count = retry_count + 1 WHERE id = $1 RETURNING retry_count
	`, id).Scan(&count)
	return count, err
}

func MarkDLQ(ctx context.Context, id string) error {
	_, err := Pool.Exec(ctx, `UPDATE submissions SET dlq = TRUE WHERE id = $1`, id)
	return err
}

func GetDLQSubmissions(ctx context.Context, limit, offset int) ([]models.Submission, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := Pool.Query(ctx, `
		SELECT s.id, s.challenge_id, c.title, s.status, s.credits_earned, s.quality_score,
		       s.extractor_version, s.approved, s.tags, s.retry_count, s.created_at
		FROM submissions s
		LEFT JOIN challenges c ON c.id = s.challenge_id
		WHERE s.dlq = TRUE ORDER BY s.created_at DESC LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Submission
	for rows.Next() {
		var s models.Submission
		var retryCount int
		if err := rows.Scan(&s.ID, &s.ChallengeID, &s.ChallengeTitle, &s.Status, &s.CreditsEarned,
			&s.QualityScore, &s.ExtractorVersion, &s.Approved, &s.Tags, &retryCount, &s.CreatedAt); err != nil {
			continue
		}
		list = append(list, s)
	}
	return list, nil
}

func ResetDLQ(ctx context.Context, id string) error {
	_, err := Pool.Exec(ctx, `UPDATE submissions SET dlq = FALSE, retry_count = 0 WHERE id = $1`, id)
	return err
}

type SubmissionRetry struct {
	models.Submission
	Robot string
}

func GetSubmissionForRetry(ctx context.Context, id string) (SubmissionRetry, error) {
	var s SubmissionRetry
	err := Pool.QueryRow(ctx, `
		SELECT s.id, s.challenge_id, c.title, s.user_id, s.video_path, s.latitude, s.longitude,
		       COALESCE(to_char(s.captured_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), ''),
		       s.robot, s.consent_version, s.video_hash
		FROM submissions s
		LEFT JOIN challenges c ON c.id = s.challenge_id
		WHERE s.id = $1
	`, id).Scan(&s.ID, &s.ChallengeID, &s.ChallengeTitle, &s.UserID, &s.VideoPath,
		&s.Latitude, &s.Longitude, &s.CapturedAt, &s.Robot, &s.ConsentVersion, &s.VideoHash)
	return s, err
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
