package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/abigailtech/humanloop/backend/middleware"
)

func GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	name := r.Context().Value(middleware.UserNameKey).(string)
	email := r.Context().Value(middleware.UserEmailKey).(string)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":          userID,
		"name":        name,
		"email":       email,
		"credits":     0,
		"submissions": 0,
	})
}
