package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/abigailtech/humanloop/backend/db"
	"github.com/abigailtech/humanloop/backend/models"
)

func CreateBuyerKey(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Label     string `json:"label"`
		DatasetID string `json:"dataset_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Label == "" {
		http.Error(w, "label required", http.StatusBadRequest)
		return
	}
	if db.Pool == nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		http.Error(w, "key generation failed", http.StatusInternalServerError)
		return
	}
	rawKey := hex.EncodeToString(raw)
	k := models.BuyerKey{
		ID:        uuid.New().String(),
		Label:     body.Label,
		DatasetID: body.DatasetID,
	}
	if err := db.CreateBuyerKey(r.Context(), k, rawKey); err != nil {
		http.Error(w, "failed to create key", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"id":         k.ID,
		"label":      k.Label,
		"dataset_id": k.DatasetID,
		"key":        rawKey,
	})
}

func ListBuyerKeys(w http.ResponseWriter, r *http.Request) {
	if db.Pool == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]any{})
		return
	}
	list, err := db.ListBuyerKeys(r.Context())
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []models.BuyerKey{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func DeleteBuyerKey(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if db.Pool == nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := db.DeleteBuyerKey(r.Context(), id); err != nil {
		http.Error(w, "failed to delete", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
