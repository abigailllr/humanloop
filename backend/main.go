package main

import (
	"log"
	"net/http"

	"github.com/abigailtech/humanloop/backend/handlers"
	"github.com/abigailtech/humanloop/backend/middleware"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/google", handlers.AuthGoogle)
	mux.HandleFunc("POST /auth/apple", handlers.AuthApple)

	mux.HandleFunc("GET /challenges", handlers.GetChallenges)

	mux.Handle("POST /submit/{challengeId}", middleware.Auth(http.HandlerFunc(handlers.Submit)))
	mux.Handle("GET /profile", middleware.Auth(http.HandlerFunc(handlers.GetProfile)))

	mux.Handle("GET /data/export", middleware.APIKey(http.HandlerFunc(handlers.ExportData)))

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
