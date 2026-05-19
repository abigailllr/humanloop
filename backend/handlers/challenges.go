package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/abigailtech/humanloop/backend/db"
	"github.com/abigailtech/humanloop/backend/models"
)

var fallbackChallenges = []models.Challenge{
	{ID: "c1", Title: "Pick & Place", Description: "Pick up any object from a table and place it into a box.", Difficulty: "Easy"},
	{ID: "c2", Title: "Fold It", Description: "Fold a piece of cloth or paper in half.", Difficulty: "Easy"},
	{ID: "c3", Title: "Sort & Stack", Description: "Sort 5 objects by size from smallest to largest.", Difficulty: "Medium"},
}

func GetChallenges(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if db.Pool != nil {
		list, err := db.GetChallenges(r.Context())
		if err == nil {
			json.NewEncoder(w).Encode(list)
			return
		}
	}

	json.NewEncoder(w).Encode(fallbackChallenges)
}
