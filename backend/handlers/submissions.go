package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/abigailtech/humanloop/backend/pipeline"
)

var Pipeline *pipeline.Pipeline

func GetSubmission(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	result, ok := Pipeline.Result(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
