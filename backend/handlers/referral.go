package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/abigailtech/humanloop/backend/db"
	"github.com/abigailtech/humanloop/backend/middleware"
)

func GetReferral(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
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
	r.Body = http.MaxBytesReader(w, r.Body, 4*1024)
	userID, ok := middleware.UserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

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
		log.Printf("redeem referral failed for %s: %v", userID, err)
		http.Error(w, "invalid referral code", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "credits_awarded": 10})
}
