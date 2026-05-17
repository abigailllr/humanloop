package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/abigailtech/humanloop/backend/models"
)

var seedChallenges = []models.Challenge{
	{ID: "c1", Title: "Pick & Place", Description: "Pick up any object from a table and place it into a box.", Submissions: 0},
	{ID: "c2", Title: "Fold It", Description: "Fold a piece of cloth or paper in half.", Submissions: 0},
	{ID: "c3", Title: "Sort & Stack", Description: "Sort 5 objects by size from smallest to largest.", Submissions: 0},
}

func GetChallenges(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(seedChallenges)
}
