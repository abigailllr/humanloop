package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/abigailtech/humanloop/backend/db"
	"github.com/abigailtech/humanloop/backend/models"
)

func GetChallenges(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if db.Pool != nil {
		list, err := db.GetChallenges(r.Context())
		if err == nil {
			json.NewEncoder(w).Encode(list)
			return
		}
	}

	json.NewEncoder(w).Encode(db.DefaultChallenges)
}

func CreateChallenge(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID          string  `json:"id"`
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Difficulty  string  `json:"difficulty"`
		ExpiresAt   *string `json:"expires_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if body.Title == "" || body.Description == "" {
		http.Error(w, "title and description required", http.StatusBadRequest)
		return
	}
	if body.ID == "" {
		b := make([]byte, 6)
		rand.Read(b)
		body.ID = "c" + hex.EncodeToString(b)
	}
	switch body.Difficulty {
	case "Easy", "Medium", "Hard":
	case "":
		body.Difficulty = "Easy"
	default:
		http.Error(w, "difficulty must be Easy, Medium, or Hard", http.StatusBadRequest)
		return
	}

	c := models.Challenge{ID: body.ID, Title: body.Title, Description: body.Description, Difficulty: body.Difficulty, ExpiresAt: body.ExpiresAt}
	if err := db.CreateChallenge(r.Context(), c); err != nil {
		http.Error(w, "failed to create", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func UpdateChallenge(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	var body struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Difficulty  string  `json:"difficulty"`
		ExpiresAt   *string `json:"expires_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if body.Difficulty != "" && body.Difficulty != "Easy" && body.Difficulty != "Medium" && body.Difficulty != "Hard" {
		http.Error(w, "difficulty must be Easy, Medium, or Hard", http.StatusBadRequest)
		return
	}
	c := models.Challenge{ID: id, Title: body.Title, Description: body.Description, Difficulty: body.Difficulty, ExpiresAt: body.ExpiresAt}
	if err := db.UpdateChallenge(r.Context(), c); err != nil {
		http.Error(w, "failed to update", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func DeleteChallenge(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	if err := db.DeleteChallenge(r.Context(), id); err != nil {
		http.Error(w, "failed to delete", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
