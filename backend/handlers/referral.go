package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/abigailtech/humanloop/backend/db"
	"github.com/abigailtech/humanloop/backend/middleware"
)

func GetReferral(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	w.Header().Set("Content-Type", "application/json")

	if db.Pool == nil {
		json.NewEncoder(w).Encode(map[string]any{"code": "", "total_referrals": 0})
		return
	}

	code, err := db.EnsureReferralCode(r.Context(), userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	stats, err := db.GetReferralStats(r.Context(), userID)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]any{"code": code, "total_referrals": 0})
		return
	}

	json.NewEncoder(w).Encode(stats)
}

func RedeemReferral(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Code == "" {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if db.Pool == nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}

	if err := db.RedeemReferral(r.Context(), userID, strings.ToUpper(body.Code)); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "credits_awarded": 10})
}
