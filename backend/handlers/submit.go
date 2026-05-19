package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/abigailtech/humanloop/backend/middleware"
	"github.com/abigailtech/humanloop/backend/models"
	"github.com/abigailtech/humanloop/backend/pipeline"
	"github.com/abigailtech/humanloop/backend/storage"
)

var Store storage.Store = storage.NewLocalStore()

func Submit(w http.ResponseWriter, r *http.Request) {
	challengeID := r.PathValue("challengeId")
	userID := r.Context().Value(middleware.UserIDKey).(string)

	file, _, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "no video", http.StatusBadRequest)
		return
	}
	defer file.Close()

	path, err := Store.Save(challengeID, userID, file)
	if err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}

	title := challengeTitle(challengeID)

	submission := models.Submission{
		ID:          uuid.New().String(),
		ChallengeID: challengeID,
		UserID:      userID,
		VideoPath:   path,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	Pipeline.Enqueue(pipeline.Job{
		SubmissionID:   submission.ID,
		ChallengeID:    submission.ChallengeID,
		ChallengeTitle: title,
		UserID:         submission.UserID,
		VideoPath:      submission.VideoPath,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"submission_id": submission.ID,
		"challenge_id":  submission.ChallengeID,
		"status":        "pending",
		"created_at":    submission.CreatedAt,
	})
}

func challengeTitle(id string) string {
	for _, c := range seedChallenges {
		if c.ID == id {
			return c.Title
		}
	}
	return ""
}
