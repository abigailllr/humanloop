package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/abigailtech/humanloop/backend/db"
	"github.com/abigailtech/humanloop/backend/metrics"
)

var startTime = time.Now()

func Health(w http.ResponseWriter, r *http.Request) {
	dbOK := db.Pool != nil && db.Pool.Ping(context.Background()) == nil

	queueDepth := int64(-1)
	queueOK := false
	if Queue != nil {
		queueDepth = Queue.Depth(r.Context())
		queueOK = true
		metrics.QueueDepth.Set(float64(queueDepth))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":      "ok",
		"uptime":      time.Since(startTime).Seconds(),
		"version":     "2.0",
		"db_ok":       dbOK,
		"queue_ok":    queueOK,
		"queue_depth": queueDepth,
	})
}
