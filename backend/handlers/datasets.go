package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/abigailtech/humanloop/backend/db"
	"github.com/abigailtech/humanloop/backend/middleware"
	"github.com/abigailtech/humanloop/backend/models"
)

func GetDatasets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if db.Pool == nil {
		json.NewEncoder(w).Encode([]any{})
		return
	}
	list, err := db.GetDatasets(r.Context())
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []models.Dataset{}
	}
	json.NewEncoder(w).Encode(list)
}

func CreateDataset(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		RobotType   string  `json:"robot_type"`
		ChallengeID string  `json:"challenge_id"`
		MinQuality  float64 `json:"min_quality"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" {
		http.Error(w, "title required", http.StatusBadRequest)
		return
	}
	d := models.Dataset{
		ID:          uuid.New().String(),
		Title:       body.Title,
		Description: body.Description,
		RobotType:   body.RobotType,
		ChallengeID: body.ChallengeID,
		MinQuality:  body.MinQuality,
	}
	if db.Pool == nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := db.CreateDataset(r.Context(), d); err != nil {
		http.Error(w, "failed to create dataset", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(d)
}

func DeleteDataset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	if db.Pool == nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := db.DeleteDataset(r.Context(), id); err != nil {
		http.Error(w, "failed to delete", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func ExportDataset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if db.Pool == nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	if scopedID, ok := r.Context().Value(middleware.BuyerDatasetKey).(string); ok && scopedID != "" && scopedID != id {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	dataset, err := db.GetDataset(r.Context(), id)
	if err != nil {
		http.Error(w, "dataset not found", http.StatusNotFound)
		return
	}
	submissions, err := db.GetDatasetSubmissions(r.Context(), dataset)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	format := r.URL.Query().Get("format")
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Content-Disposition", `attachment; filename="`+strings.ReplaceAll(dataset.Title, " ", "_")+`.ndjson"`)

	for _, s := range submissions {
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

func GetDatasetStats(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if db.Pool == nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	dataset, err := db.GetDataset(r.Context(), id)
	if err != nil {
		http.Error(w, "dataset not found", http.StatusNotFound)
		return
	}
	stats, err := db.GetDatasetStats(r.Context(), dataset)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	stats["dataset_id"] = id
	stats["title"] = dataset.Title
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func GetDatasetSubmissionList(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if db.Pool == nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	dataset, err := db.GetDataset(r.Context(), id)
	if err != nil {
		http.Error(w, "dataset not found", http.StatusNotFound)
		return
	}
	submissions, err := db.GetDatasetSubmissions(r.Context(), dataset)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if submissions == nil {
		submissions = []models.Submission{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"dataset":     dataset,
		"submissions": submissions,
		"total":       len(submissions),
	})
}
