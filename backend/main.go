package main

import (
	"log"
	"net/http"
	"os"

	"github.com/abigailtech/humanloop/backend/handlers"
	"github.com/abigailtech/humanloop/backend/middleware"
	"github.com/abigailtech/humanloop/backend/pipeline"
	"github.com/abigailtech/humanloop/backend/storage"
)

func main() {
	extractedDir := os.Getenv("EXTRACTED_DIR")
	if extractedDir == "" {
		extractedDir = "./data/extracted"
	}

	handlers.Pipeline = pipeline.New(4, extractedDir)

	if os.Getenv("AWS_BUCKET") != "" {
		s3, err := storage.NewS3Store()
		if err != nil {
			log.Fatal("s3 init failed:", err)
		}
		handlers.Store = s3
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/google", handlers.AuthGoogle)
	mux.HandleFunc("POST /auth/apple", handlers.AuthApple)

	mux.HandleFunc("GET /challenges", handlers.GetChallenges)

	mux.Handle("POST /submit/{challengeId}", middleware.Auth(http.HandlerFunc(handlers.Submit)))
	mux.Handle("GET /profile", middleware.Auth(http.HandlerFunc(handlers.GetProfile)))
	mux.Handle("GET /submissions/{id}", middleware.Auth(http.HandlerFunc(handlers.GetSubmission)))

	mux.Handle("GET /data/export", middleware.APIKey(http.HandlerFunc(handlers.ExportData)))

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

