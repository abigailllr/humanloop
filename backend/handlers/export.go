package handlers

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func ExportData(w http.ResponseWriter, r *http.Request) {
	dir := os.Getenv("DATA_DIR")
	if dir == "" {
		dir = "./data/extracted"
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]any{})
		return
	}

	challengeFilter := r.URL.Query().Get("challenge_id")
	difficultyFilter := strings.ToLower(r.URL.Query().Get("difficulty"))
	minConf, _ := strconv.ParseFloat(r.URL.Query().Get("min_confidence"), 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	format := r.URL.Query().Get("format")

	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	var records []map[string]any
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".gz" {
			continue
		}
		record, err := readHMDF(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		if challengeFilter != "" && record["challenge_id"] != challengeFilter {
			continue
		}
		if difficultyFilter != "" {
			if diff, ok := record["difficulty"].(string); !ok || strings.ToLower(diff) != difficultyFilter {
				continue
			}
		}
		if minConf > 0 {
			if val, ok := record["validation"].(map[string]any); ok {
				conf, _ := val["confidence"].(float64)
				if conf < minConf {
					continue
				}
			}
		}
		records = append(records, record)
	}

	if offset >= len(records) {
		records = nil
	} else {
		records = records[offset:]
		if len(records) > limit {
			records = records[:limit]
		}
	}

	w.Header().Set("Content-Type", "application/json")

	switch format {
	case "ndjson":
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.Header().Set("Content-Disposition", "attachment; filename=\"humanloop-export.ndjson\"")
		for _, rec := range records {
			line, _ := json.Marshal(rec)
			w.Write(line)
			w.Write([]byte("\n"))
		}
	case "lerobot":
		w.Header().Set("Content-Disposition", "attachment; filename=\"humanloop-lerobot.ndjson\"")
		for _, rec := range records {
			row := lerobotRow(rec)
			line, _ := json.Marshal(row)
			w.Write(line)
			w.Write([]byte("\n"))
		}
	default:
		json.NewEncoder(w).Encode(records)
	}
}

func ExportStats(w http.ResponseWriter, r *http.Request) {
	dir := os.Getenv("DATA_DIR")
	if dir == "" {
		dir = "./data/extracted"
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"total": 0})
		return
	}

	total := 0
	totalFrames := 0
	byChallengeID := map[string]int{}

	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".gz" {
			continue
		}
		record, err := readHMDF(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		total++
		if frames, ok := record["frames"].([]any); ok {
			totalFrames += len(frames)
		}
		if cid, ok := record["challenge_id"].(string); ok && cid != "" {
			byChallengeID[cid]++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"total_episodes":  total,
		"total_frames":    totalFrames,
		"by_challenge_id": byChallengeID,
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
