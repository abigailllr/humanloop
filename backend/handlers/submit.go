package handlers

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"time"

	"github.com/google/uuid"

	"github.com/abigailtech/humanloop/backend/middleware"
	"github.com/abigailtech/humanloop/backend/models"
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

	submission := models.Submission{
		ID:          uuid.New().String(),
		ChallengeID: challengeID,
		UserID:      userID,
		VideoPath:   path,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	validation := runExtractor("validate", path)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"submission_id": submission.ID,
		"challenge_id":  submission.ChallengeID,
		"user_id":       submission.UserID,
		"video_path":    submission.VideoPath,
		"created_at":    submission.CreatedAt,
		"validation":    validation,
	})
}

func runExtractor(command, videoPath string) map[string]any {
	out, err := exec.Command("python3", "../extractor/main.py", command, videoPath).Output()
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	var result map[string]any
	json.Unmarshal(out, &result)
	return result
}
