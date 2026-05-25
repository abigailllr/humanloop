package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/abigailtech/humanloop/backend/db"
	"github.com/abigailtech/humanloop/backend/hub"
	"github.com/abigailtech/humanloop/backend/middleware"
)

func StreamSubmission(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	userID := r.Context().Value(middleware.UserIDKey).(string)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	sendEvent := func(data map[string]any) {
		b, _ := json.Marshal(data)
		fmt.Fprintf(w, "data: %s\n\n", b)
		flusher.Flush()
	}

	isTerminal := func(status string) bool {
		return status == "done" || status == "failed" || status == "synthetic"
	}

	if db.Pool != nil {
		s, err := db.GetSubmission(r.Context(), id)
		if err == nil && s.UserID == userID {
			sendEvent(map[string]any{"submission_id": id, "status": s.Status})
			if isTerminal(s.Status) {
				return
			}
		}
	} else {
		if result, ok := Pipeline.Result(id); ok {
			sendEvent(map[string]any{"submission_id": id, "status": string(result.Status)})
			if isTerminal(string(result.Status)) {
				return
			}
		} else {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}

	ch := hub.Default.Subscribe(id)
	defer hub.Default.Unsubscribe(id, ch)

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
			var ev map[string]any
			if json.Unmarshal(msg, &ev) == nil {
				if st, _ := ev["status"].(string); isTerminal(st) {
					return
				}
			}
		}
	}
}
