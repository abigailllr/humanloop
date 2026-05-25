package db

import "context"

func GetAnalytics(ctx context.Context) (map[string]any, error) {
	rows, err := Pool.Query(ctx, `
		SELECT DATE(created_at) AS day, COUNT(*)
		FROM submissions
		WHERE status = 'done' AND created_at > NOW() - INTERVAL '30 days'
		GROUP BY day ORDER BY day
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	perDay := []map[string]any{}
	for rows.Next() {
		var day string
		var count int
		if rows.Scan(&day, &count) == nil {
			perDay = append(perDay, map[string]any{"day": day, "count": count})
		}
	}

	robotRows, err := Pool.Query(ctx, `
		SELECT robot, COUNT(*) FROM submissions WHERE status = 'done' GROUP BY robot
	`)
	if err != nil {
		return nil, err
	}
	defer robotRows.Close()
	byRobot := map[string]int{}
	for robotRows.Next() {
		var robot string
		var count int
		if robotRows.Scan(&robot, &count) == nil {
			byRobot[robot] = count
		}
	}

	challengeRows, err := Pool.Query(ctx, `
		SELECT challenge_id, COUNT(*) FROM submissions WHERE status = 'done' GROUP BY challenge_id
	`)
	if err != nil {
		return nil, err
	}
	defer challengeRows.Close()
	byChallenge := map[string]int{}
	for challengeRows.Next() {
		var id string
		var count int
		if challengeRows.Scan(&id, &count) == nil {
			byChallenge[id] = count
		}
	}

	buckets := []map[string]any{}
	boundaries := [][2]float64{{0, 0.2}, {0.2, 0.4}, {0.4, 0.6}, {0.6, 0.8}, {0.8, 1.01}}
	for _, b := range boundaries {
		var count int
		Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM submissions WHERE status='done' AND quality_score >= $1 AND quality_score < $2
		`, b[0], b[1]).Scan(&count)
		buckets = append(buckets, map[string]any{"min": b[0], "max": b[1], "count": count})
	}

	return map[string]any{
		"submissions_per_day":    perDay,
		"by_robot":               byRobot,
		"by_challenge":           byChallenge,
		"quality_distribution":   buckets,
	}, nil
}
