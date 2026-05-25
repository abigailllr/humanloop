package handlers

import (
	"net/http"

	"github.com/abigailtech/humanloop/backend/db"
	"github.com/abigailtech/humanloop/backend/middleware"
)

func DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	if db.Pool == nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := db.DeleteUser(r.Context(), userID); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
