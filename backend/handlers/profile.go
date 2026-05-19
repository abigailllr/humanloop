package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/abigailtech/humanloop/backend/db"
	"github.com/abigailtech/humanloop/backend/middleware"
)

func GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	name := r.Context().Value(middleware.UserNameKey).(string)
	email := r.Context().Value(middleware.UserEmailKey).(string)

	w.Header().Set("Content-Type", "application/json")

	if db.Pool != nil {
		u, err := db.GetUser(r.Context(), userID)
		if err == nil {
			json.NewEncoder(w).Encode(u)
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]any{
		"id":          userID,
		"name":        name,
		"email":       email,
		"credits":     0,
		"submissions": 0,
	})
}
