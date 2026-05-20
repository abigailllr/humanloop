package handlers

import (
	"context"
	"encoding/json"
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

	title := challengeTitle(challengeID)

	submission := models.Submission{
		ID:          uuid.New().String(),
		ChallengeID: challengeID,
		UserID:      userID,
		VideoPath:   path,
		Latitude:    lat,
		Longitude:   lng,
		CapturedAt:  capturedAt,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	job := pipeline.Job{
		SubmissionID:   submission.ID,
		ChallengeID:    submission.ChallengeID,
		ChallengeTitle: title,
		UserID:         submission.UserID,
		VideoPath:      submission.VideoPath,
		Latitude:       lat,
		Longitude:      lng,
		CapturedAt:     capturedAt,
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
		challenges, err := db.GetChallenges(context.Background())
		if err == nil {
			for _, c := range challenges {
				if c.ID == id {
					return c.Title
				}
			}
		}
	}
	for _, c := range db.DefaultChallenges {
		if c.ID == id {
			return c.Title
		}
	}
	return ""
}
