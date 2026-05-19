package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/abigailtech/humanloop/backend/db"
	"github.com/abigailtech/humanloop/backend/handlers"
	"github.com/abigailtech/humanloop/backend/middleware"
	"github.com/abigailtech/humanloop/backend/pipeline"
	"github.com/abigailtech/humanloop/backend/queue"
	"github.com/abigailtech/humanloop/backend/storage"
)

func main() {
	ctx := context.Background()

	if os.Getenv("JWT_SECRET") == "" {
		log.Println("warning: JWT_SECRET not set, using insecure default")
	}

	extractedDir := os.Getenv("EXTRACTED_DIR")
	if extractedDir == "" {
		extractedDir = "./data/extracted"
	}

	handlers.Pipeline = pipeline.New(4, extractedDir)

	if err := db.Connect(ctx); err != nil {
		log.Println("db unavailable:", err)
	} else {
		defer db.Close()
		if err := db.Migrate(ctx); err != nil {
			log.Println("db migrate:", err)
		} else {
			if err := db.SeedChallenges(ctx); err != nil {
				log.Println("db seed:", err)
			}
		}
	}

	q := queue.New()
	if err := q.Ping(ctx); err != nil {
		log.Println("redis unavailable:", err)
	} else {
		handlers.Queue = q
		go func() {
			for {
				data, err := q.Pop(ctx)
				if err != nil {
					continue
				}
				var job pipeline.Job
				if err := json.Unmarshal(data, &job); err != nil {
					continue
				}
				handlers.Pipeline.Enqueue(job)
			}
		}()
	}

	if os.Getenv("AWS_BUCKET") != "" {
		s3, err := storage.NewS3Store()
		if err != nil {
			log.Fatal("s3 init failed:", err)
		}
		handlers.Store = s3
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.Health)
	mux.Handle("GET /metrics", middleware.APIKey(http.HandlerFunc(handlers.Metrics)))

	authRL := middleware.RateLimitAuth
	submitRL := middleware.RateLimitSubmit

	mux.Handle("POST /v1/auth/google", authRL(http.HandlerFunc(handlers.AuthGoogle)))
	mux.Handle("POST /v1/auth/apple", authRL(http.HandlerFunc(handlers.AuthApple)))

	mux.HandleFunc("GET /v1/challenges", handlers.GetChallenges)

	mux.Handle("POST /v1/submit/{challengeId}", submitRL(middleware.Auth(http.HandlerFunc(handlers.Submit))))
	mux.Handle("GET /v1/profile", middleware.Auth(http.HandlerFunc(handlers.GetProfile)))
	mux.Handle("GET /v1/submissions/{id}", middleware.Auth(http.HandlerFunc(handlers.GetSubmission)))

	mux.HandleFunc("GET /v1/leaderboard", handlers.GetLeaderboard)
	mux.Handle("GET /v1/submissions", middleware.Auth(http.HandlerFunc(handlers.GetUserSubmissions)))

	mux.Handle("GET /v1/data/export", middleware.APIKey(http.HandlerFunc(handlers.ExportData)))

	stack := middleware.Security(middleware.RealIP(middleware.Logger(mux)))
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", stack))
}
