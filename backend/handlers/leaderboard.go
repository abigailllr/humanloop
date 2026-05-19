package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/abigailtech/humanloop/backend/db"
)

func GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if db.Pool == nil {
		json.NewEncoder(w).Encode([]any{})
		return
	}

	entries, err := db.GetLeaderboard(r.Context())
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

	result := make([]entry, len(entries))
	for i, u := range entries {
		result[i] = entry{Rank: i + 1, ID: u.ID, Name: u.Name, Credits: u.Credits, Submissions: u.Submissions}
	}

	json.NewEncoder(w).Encode(result)
}
