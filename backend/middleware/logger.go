package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)
		entry, _ := json.Marshal(map[string]any{
			"method":   r.Method,
			"path":     r.URL.Path,
			"status":   rw.status,
			"duration": time.Since(start).Milliseconds(),
			"ip":       realIP(r),
		})
		log.Println(string(entry))
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}
