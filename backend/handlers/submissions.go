package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/abigailtech/humanloop/backend/db"
	"github.com/abigailtech/humanloop/backend/middleware"
	"github.com/abigailtech/humanloop/backend/models"
	"github.com/abigailtech/humanloop/backend/pipeline"
)

var Pipeline *pipeline.Pipeline

func GetSubmission(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	userID := r.Context().Value(middleware.UserIDKey).(string)
	w.Header().Set("Content-Type", "application/json")

	if result, ok := Pipeline.Result(id); ok {
		if result.SubmissionID != id {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(result)
		return
	}

	if db.Pool != nil {
		s, err := db.GetSubmission(r.Context(), id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if s.UserID != userID {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"submission_id": s.ID,
			"status":        s.Status,
		})
		return
	}

	http.Error(w, "not found", http.StatusNotFound)
}

func GetUserSubmissions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	w.Header().Set("Content-Type", "application/json")

	if db.Pool == nil {
		json.NewEncoder(w).Encode([]any{})
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	list, err := db.GetSubmissionsByUser(r.Context(), userID, limit, offset)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	if list == nil {
		list = []models.Submission{}
	}
	json.NewEncoder(w).Encode(list)
}
