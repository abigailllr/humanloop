package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/abigailtech/humanloop/backend/db"
	"github.com/abigailtech/humanloop/backend/middleware"
	"github.com/abigailtech/humanloop/backend/pipeline"
)

func GetChallengeStats(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if db.Pool == nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	stats, err := db.GetChallengeStats(r.Context(), id)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func GetChallengeLeaderboard(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if db.Pool == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]any{})
		return
	}
	list, err := db.GetChallengeLeaderboard(r.Context(), id, limit)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	type entry struct {
		Rank        int    `json:"rank"`
		ID          string `json:"id"`
		Name        string `json:"name"`
		Credits     int    `json:"credits"`
		Submissions int    `json:"submissions"`
	}
	result := make([]entry, len(list))
	for i, u := range list {
		result[i] = entry{Rank: i + 1, ID: u.ID, Name: u.Name, Credits: u.Credits, Submissions: u.Submissions}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func GetUserStats(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	if db.Pool == nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	stats, err := db.GetUserStats(r.Context(), userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func GetCreditHistory(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if db.Pool == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]any{})
		return
	}
	history, err := db.GetCreditHistory(r.Context(), userID, limit, offset)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if history == nil {
		history = []map[string]any{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func GetHeatmap(w http.ResponseWriter, r *http.Request) {
	if db.Pool == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]any{})
		return
	}
	points, err := db.GetSubmissionHeatmap(r.Context())
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if points == nil {
		points = []map[string]any{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(points)
}

func AdminRetrySubmission(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if db.Pool == nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	s, err := db.GetSubmissionForRetry(r.Context(), id)
	if err != nil {
		http.Error(w, "submission not found", http.StatusNotFound)
		return
	}
	job := pipeline.Job{
		SubmissionID:   s.ID,
		ChallengeID:    s.ChallengeID,
		ChallengeTitle: s.ChallengeTitle,
		UserID:         s.UserID,
		VideoPath:      s.VideoPath,
		Latitude:       s.Latitude,
		Longitude:      s.Longitude,
		CapturedAt:     s.CapturedAt,
		Robot:          s.Robot,
		VideoHash:      s.VideoHash,
		ConsentVersion: s.ConsentVersion,
	}
	db.UpdateSubmissionStatus(r.Context(), id, "pending", "", 0)
	Pipeline.Enqueue(job)
	w.WriteHeader(http.StatusAccepted)
}
