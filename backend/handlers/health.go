package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

var startTime = time.Now()

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "ok",
		"uptime":  time.Since(startTime).Seconds(),
		"version": "1.2",
	})
}
