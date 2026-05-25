package handlers

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/abigailtech/humanloop/backend/db"
)

func ExportData(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	challengeFilter := q.Get("challenge_id")
	robotFilter := q.Get("robot")
	minQuality, _ := strconv.ParseFloat(q.Get("min_quality"), 64)
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	format := q.Get("format")

	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	approved := true
	submissions, err := db.GetAdminSubmissions(r.Context(), "done", robotFilter, &approved, limit, offset)
	if err != nil || db.Pool == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]any{})
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")

	for _, s := range submissions {
		if challengeFilter != "" && s.ChallengeID != challengeFilter {
			continue
		}
		if minQuality > 0 && s.QualityScore < minQuality {
			continue
		}
		if s.HmdfPath == "" {
			continue
		}
		record, err := readHMDF(s.HmdfPath)
		if err != nil {
			continue
		}
		var line []byte
		if format == "lerobot" {
			line, _ = json.Marshal(lerobotRow(record))
		} else {
			line, _ = json.Marshal(record)
		}
		w.Write(line)
		w.Write([]byte("\n"))
	}
}

func ExportStats(w http.ResponseWriter, r *http.Request) {
	if db.Pool == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"total_episodes": 0})
		return
	}
	approved := true
	submissions, err := db.GetAdminSubmissions(r.Context(), "done", "", &approved, 10000, 0)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	total := len(submissions)
	byChallengeID := map[string]int{}
	byRobot := map[string]int{}
	var qualitySum float64
	for _, s := range submissions {
		byChallengeID[s.ChallengeID]++
		qualitySum += s.QualityScore
	}
	avgQ := 0.0
	if total > 0 {
		avgQ = qualitySum / float64(total)
	}
	for _, s := range submissions {
		byRobot[s.Robot]++
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"total_episodes":  total,
		"quality_avg":     avgQ,
		"by_challenge_id": byChallengeID,
		"by_robot":        byRobot,
	})
}

func lerobotRow(record map[string]any) map[string]any {
	frames, _ := record["frames"].([]any)
	rows := []map[string]any{}
	for i, f := range frames {
		frame, ok := f.(map[string]any)
		if !ok {
			continue
		}
		ms, _ := frame["motor_state"].(map[string]any)
		obs, _ := frame["obs"].([]any)
		q, _ := ms["q"].([]any)
		var nextQ []any
		if i+1 < len(frames) {
			if nf, ok := frames[i+1].(map[string]any); ok {
				if nms, ok := nf["motor_state"].(map[string]any); ok {
					nextQ, _ = nms["q"].([]any)
				}
			}
		}
		if nextQ == nil {
			nextQ = q
		}
		t, _ := frame["t"].(float64)
		rows = append(rows, map[string]any{
			"frame_index":       i,
			"timestamp":         t,
			"observation.state": obs,
			"action":            nextQ,
			"motor_state.q":     q,
		})
	}
	return map[string]any{
		"challenge_id":  record["challenge_id"],
		"submission_id": record["submission_id"],
		"user_id":       record["user_id"],
		"fps":           record["fps"],
		"frames":        rows,
	}
}

func readHMDF(path string) (map[string]any, error) {
	if strings.HasPrefix(path, "s3://") {
		return nil, fmt.Errorf("s3 hmdf not available for local export")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	b, err := io.ReadAll(gz)
	if err != nil {
		return nil, err
	}
	var record map[string]any
	if err := json.Unmarshal(b, &record); err != nil {
		return nil, err
	}
	return record, nil
}
