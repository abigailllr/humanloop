package main

import (
	"log"
	"net/http"

	"github.com/abigailtech/humanloop/backend/handlers"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /challenges", handlers.GetChallenges)
	mux.HandleFunc("POST /submit/{challengeId}", handlers.Submit)
	mux.HandleFunc("GET /profile", handlers.GetProfile)
	mux.HandleFunc("GET /data/export", handlers.ExportData)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
