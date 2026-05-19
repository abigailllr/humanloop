package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"compress/gzip"
	"io"
)

func ExportData(w http.ResponseWriter, r *http.Request) {
	dir := os.Getenv("DATA_DIR")
	if dir == "" {
		dir = "./data/extracted"
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]any{})
		return
	}

	format := r.URL.Query().Get("format")

	if format == "ndjson" {
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.Header().Set("Content-Disposition", "attachment; filename=\"humanloop-export.ndjson\"")
		for _, e := range entries {
			if filepath.Ext(e.Name()) != ".gz" {
				continue
			}
			record, err := readHMDF(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			line, _ := json.Marshal(record)
			w.Write(line)
			w.Write([]byte("\n"))
		}
		return
	}

	var records []map[string]any
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".gz" {
			continue
		}
		record, err := readHMDF(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		records = append(records, record)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

func readHMDF(path string) (map[string]any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	b, err := io.ReadAll(gz)
	if err != nil {
		return nil, err
	}

	var record map[string]any
	if err := json.Unmarshal(b, &record); err != nil {
		return nil, err
	}
	return record, nil
}
