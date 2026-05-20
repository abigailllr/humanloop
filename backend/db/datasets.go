package db

import (
	"context"
	"fmt"

	"github.com/abigailtech/humanloop/backend/models"
)

func CreateDataset(ctx context.Context, d models.Dataset) error {
	_, err := Pool.Exec(ctx, `
		INSERT INTO datasets (id, title, description, robot_type, challenge_id, min_quality)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, d.ID, d.Title, d.Description, d.RobotType, d.ChallengeID, d.MinQuality)
	return err
}

func GetDatasets(ctx context.Context) ([]models.Dataset, error) {
	rows, err := Pool.Query(ctx, `SELECT id, title, description, robot_type, challenge_id, min_quality, created_at FROM datasets ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Dataset
	for rows.Next() {
		var d models.Dataset
		if err := rows.Scan(&d.ID, &d.Title, &d.Description, &d.RobotType, &d.ChallengeID, &d.MinQuality, &d.CreatedAt); err != nil {
			continue
		}
		list = append(list, d)
	}
	return list, nil
}

func GetDataset(ctx context.Context, id string) (models.Dataset, error) {
	var d models.Dataset
	err := Pool.QueryRow(ctx, `SELECT id, title, description, robot_type, challenge_id, min_quality, created_at FROM datasets WHERE id = $1`, id).
		Scan(&d.ID, &d.Title, &d.Description, &d.RobotType, &d.ChallengeID, &d.MinQuality, &d.CreatedAt)
	return d, err
}

func DeleteDataset(ctx context.Context, id string) error {
	_, err := Pool.Exec(ctx, `DELETE FROM datasets WHERE id = $1`, id)
	return err
}

func GetDatasetSubmissions(ctx context.Context, d models.Dataset) ([]models.Submission, error) {
	args := []any{true, "done"}
	where := "s.approved = $1 AND s.status = $2"
	i := 3
	if d.RobotType != "" {
		where += fmt.Sprintf(" AND s.robot = $%d", i)
		args = append(args, d.RobotType)
		i++
	}
	if d.ChallengeID != "" {
		where += fmt.Sprintf(" AND s.challenge_id = $%d", i)
		args = append(args, d.ChallengeID)
		i++
	}
	if d.MinQuality > 0 {
		where += fmt.Sprintf(" AND s.quality_score >= $%d", i)
		args = append(args, d.MinQuality)
		i++
	}
	rows, err := Pool.Query(ctx, fmt.Sprintf(`
		SELECT s.id, s.challenge_id, s.status, s.hmdf_path, s.quality_score, s.extractor_version, s.tags, s.created_at
		FROM submissions s WHERE %s ORDER BY s.quality_score DESC
	`, where), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Submission
	for rows.Next() {
		var s models.Submission
		if err := rows.Scan(&s.ID, &s.ChallengeID, &s.Status, &s.HmdfPath, &s.QualityScore, &s.ExtractorVersion, &s.Tags, &s.CreatedAt); err != nil {
			continue
		}
		list = append(list, s)
	}
	return list, nil
}

func GetDatasetStats(ctx context.Context, d models.Dataset) (map[string]any, error) {
	args := []any{true, "done"}
	where := "approved = $1 AND status = $2"
	i := 3
	if d.RobotType != "" {
		where += fmt.Sprintf(" AND robot = $%d", i)
		args = append(args, d.RobotType)
		i++
	}
	if d.ChallengeID != "" {
		where += fmt.Sprintf(" AND challenge_id = $%d", i)
		args = append(args, d.ChallengeID)
		i++
	}
	if d.MinQuality > 0 {
		where += fmt.Sprintf(" AND quality_score >= $%d", i)
		args = append(args, d.MinQuality)
		i++
	}
	var count int
	var avgQ, minQ, maxQ float64
	err := Pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*), COALESCE(AVG(quality_score),0), COALESCE(MIN(quality_score),0), COALESCE(MAX(quality_score),0)
		FROM submissions WHERE %s
	`, where), args...).Scan(&count, &avgQ, &minQ, &maxQ)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"submission_count": count,
		"quality_avg":      avgQ,
		"quality_min":      minQ,
		"quality_max":      maxQ,
	}, nil
}
