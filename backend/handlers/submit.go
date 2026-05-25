package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/abigailtech/humanloop/backend/db"
	"github.com/abigailtech/humanloop/backend/middleware"
	"github.com/abigailtech/humanloop/backend/models"
	"github.com/abigailtech/humanloop/backend/pipeline"
	"github.com/abigailtech/humanloop/backend/queue"
	"github.com/abigailtech/humanloop/backend/storage"
)

var Store storage.Store = storage.NewLocalStore()
var Queue *queue.Client

const maxVideoSize = 500 * 1024 * 1024

func Submit(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxVideoSize+4*1024)

	challengeID := r.PathValue("challengeId")
	userID := r.Context().Value(middleware.UserIDKey).(string)

	if db.Pool != nil {
		if count, err := db.CountSubmissionsToday(r.Context(), userID); err == nil && count >= 20 {
			http.Error(w, "daily submission limit reached", http.StatusTooManyRequests)
			return
		}
	}

	if err := r.ParseMultipartForm(32 * 1024 * 1024); err != nil {
		http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "no video", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Size > maxVideoSize {
		http.Error(w, "video too large", http.StatusRequestEntityTooLarge)
		return
	}

	sniff := make([]byte, 512)
	if _, err := file.Read(sniff); err != nil {
		http.Error(w, "unreadable file", http.StatusBadRequest)
		return
	}
	mime := http.DetectContentType(sniff)
	if mime != "video/mp4" && mime != "video/quicktime" && mime != "video/x-msvideo" && mime != "video/webm" && mime != "video/x-matroska" {
		http.Error(w, "invalid file type", http.StatusUnsupportedMediaType)
		return
	}
	if _, err := file.Seek(0, 0); err != nil {
		http.Error(w, "unreadable file", http.StatusBadRequest)
		return
	}

	hasher := sha256.New()
	if _, err := io.Copy(hasher, io.LimitReader(file, 4*1024*1024)); err != nil {
		http.Error(w, "unreadable file", http.StatusBadRequest)
		return
	}
	videoHash := hex.EncodeToString(hasher.Sum(nil))
	if _, err := file.Seek(0, 0); err != nil {
		http.Error(w, "unreadable file", http.StatusBadRequest)
		return
	}

	if db.Pool != nil {
		if exists, err := db.VideoHashExists(r.Context(), videoHash); err == nil && exists {
			http.Error(w, "duplicate submission", http.StatusConflict)
			return
		}
	}

	path, err := Store.Save(challengeID, userID, file)
	if err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}

	lat, _ := strconv.ParseFloat(r.FormValue("lat"), 64)
	lng, _ := strconv.ParseFloat(r.FormValue("lng"), 64)
	capturedAt := r.FormValue("captured_at")
	if capturedAt == "" {
		capturedAt = time.Now().UTC().Format(time.RFC3339)
	} else if _, err := time.Parse(time.RFC3339, capturedAt); err != nil {
		capturedAt = time.Now().UTC().Format(time.RFC3339)
	}
	robot := r.FormValue("robot")
	if robot == "" {
		robot = "generic_bimanual"
	}
	consentVersion := r.FormValue("consent_version")
	if consentVersion == "" {
		consentVersion = "1.0"
	}

	if db.Pool != nil {
		if _, err := db.GetChallenge(r.Context(), challengeID); err != nil {
			http.Error(w, "challenge not found", http.StatusNotFound)
			return
		}
	}

	title := challengeTitle(challengeID)
	difficulty := challengeDifficulty(challengeID)
	userEmail := r.Context().Value(middleware.UserEmailKey).(string)
	userName := r.Context().Value(middleware.UserNameKey).(string)

	submission := models.Submission{
		ID:             uuid.New().String(),
		ChallengeID:    challengeID,
		UserID:         userID,
		VideoPath:      path,
		Latitude:       lat,
		Longitude:      lng,
		CapturedAt:     capturedAt,
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
		ConsentVersion: consentVersion,
		VideoHash:      videoHash,
	}

	job := pipeline.Job{
		SubmissionID:        submission.ID,
		ChallengeID:         submission.ChallengeID,
		ChallengeTitle:      title,
		ChallengeDifficulty: difficulty,
		UserID:              submission.UserID,
		UserEmail:           userEmail,
		UserName:            userName,
		VideoPath:           submission.VideoPath,
		Latitude:            lat,
		Longitude:           lng,
		CapturedAt:          capturedAt,
		Robot:               robot,
		VideoHash:           videoHash,
		ConsentVersion:      consentVersion,
	}

	if db.Pool != nil {
		db.CreateSubmission(r.Context(), submission)
	}

	if Queue != nil {
		if err := Queue.Push(r.Context(), job); err != nil {
			log.Println("queue push failed:", err)
			Pipeline.Enqueue(job)
		}
	} else {
		Pipeline.Enqueue(job)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"submission_id": submission.ID,
		"challenge_id":  submission.ChallengeID,
		"status":        "pending",
		"created_at":    submission.CreatedAt,
	})
}

func challengeTitle(id string) string {
	if db.Pool != nil {
		if c, err := db.GetChallenge(context.Background(), id); err == nil {
			return c.Title
		}
	}
	for _, c := range db.DefaultChallenges {
		if c.ID == id {
			return c.Title
		}
	}
	return ""
}

func challengeDifficulty(id string) string {
	if db.Pool != nil {
		if c, err := db.GetChallenge(context.Background(), id); err == nil {
			return c.Difficulty
		}
	}
	for _, c := range db.DefaultChallenges {
		if c.ID == id {
			return c.Difficulty
		}
	}
	return "Easy"
}
