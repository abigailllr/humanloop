package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/abigailtech/humanloop/backend/db"
	"github.com/abigailtech/humanloop/backend/middleware"
	"github.com/abigailtech/humanloop/backend/pipeline"
)

var Pipeline *pipeline.Pipeline

func GetSubmission(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	result, ok := Pipeline.Result(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func GetUserSubmissions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	w.Header().Set("Content-Type", "application/json")

	if db.Pool == nil {
		json.NewEncoder(w).Encode([]any{})
		return
	}

	list, err := db.GetSubmissionsByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(list)
}
