package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/abigailtech/humanloop/backend/db"
	"github.com/abigailtech/humanloop/backend/models"
)

func AdminGetSubmissions(w http.ResponseWriter, r *http.Request) {
	if db.Pool == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]any{})
		return
	}
	q := r.URL.Query()
	status := q.Get("status")
	robot := q.Get("robot")
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	if limit <= 0 || limit > 500 {
		limit = 50
	}

	var approved *bool
	if v := q.Get("approved"); v == "true" {
		t := true
		approved = &t
	} else if v == "false" {
		f := false
		approved = &f
	}

	list, err := db.GetAdminSubmissions(r.Context(), status, robot, approved, limit, offset)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = make([]models.Submission, 0)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func AdminApproveSubmission(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if db.Pool == nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := db.ApproveSubmission(r.Context(), id, true); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func AdminGetDLQ(w http.ResponseWriter, r *http.Request) {
	if db.Pool == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]any{})
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	list, err := db.GetDLQSubmissions(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = make([]models.Submission, 0)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func AdminRetryDLQ(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if db.Pool == nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := db.ResetDLQ(r.Context(), id); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	AdminRetrySubmission(w, r)
}

func AdminRejectSubmission(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if db.Pool == nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := db.ApproveSubmission(r.Context(), id, false); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func AdminTagSubmission(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if db.Pool == nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := db.TagSubmission(r.Context(), id, body.Tags); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
