package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/abigailtech/humanloop/backend/db"
	"github.com/abigailtech/humanloop/backend/middleware"
)

func userLevel(doneSubmissions int) string {
	switch {
	case doneSubmissions >= 100:
		return "Elite"
	case doneSubmissions >= 50:
		return "Master"
	case doneSubmissions >= 20:
		return "Expert"
	case doneSubmissions >= 5:
		return "Explorer"
	default:
		return "Rookie"
	}
}

func computeBadges(doneCount int, avgQuality float64, totalCredits int) []string {
	var badges []string
	if totalCredits > 0 || doneCount > 0 {
		badges = append(badges, "first_submission")
	}
	if doneCount >= 1 {
		badges = append(badges, "first_verified")
	}
	if doneCount >= 10 {
		badges = append(badges, "ten_verified")
	}
	if doneCount >= 50 {
		badges = append(badges, "fifty_verified")
	}
	if doneCount >= 100 {
		badges = append(badges, "century")
	}
	if doneCount >= 5 && avgQuality >= 0.8 {
		badges = append(badges, "quality_star")
	}
	return badges
}

func GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	name := r.Context().Value(middleware.UserNameKey).(string)
	email := r.Context().Value(middleware.UserEmailKey).(string)

	w.Header().Set("Content-Type", "application/json")

	if db.Pool != nil {
		u, err := db.GetUser(r.Context(), userID)
		if err == nil {
			var doneCount int
			var avgQuality float64
			db.Pool.QueryRow(r.Context(), `
				SELECT COUNT(*), COALESCE(AVG(quality_score),0)
				FROM submissions WHERE user_id=$1 AND status='done'
			`, userID).Scan(&doneCount, &avgQuality)

			level := userLevel(doneCount)
			badges := computeBadges(doneCount, avgQuality, u.Credits)

			if u.ReferralCode == "" {
				u.ReferralCode, _ = db.EnsureReferralCode(r.Context(), userID)
			}

			json.NewEncoder(w).Encode(map[string]any{
				"id":            u.ID,
				"email":         u.Email,
				"name":          u.Name,
				"credits":       u.Credits,
				"submissions":   u.Submissions,
				"referral_code": u.ReferralCode,
				"level":         level,
				"badges":        badges,
			})
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]any{
		"id":          userID,
		"name":        name,
		"email":       email,
		"credits":     0,
		"submissions": 0,
		"level":       "Rookie",
		"badges":      []string{},
	})
}
