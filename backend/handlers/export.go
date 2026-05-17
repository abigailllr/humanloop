package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

func ExportData(w http.ResponseWriter, r *http.Request) {
	dir := os.Getenv("DATA_DIR")
	if dir == "" {
		dir = "./data/extracted"
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		json.NewEncoder(w).Encode([]any{})
		return
	}

	var records []map[string]any
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var rec map[string]any
		if json.Unmarshal(b, &rec) == nil {
			records = append(records, rec)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}
