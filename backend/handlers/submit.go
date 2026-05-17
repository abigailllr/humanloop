package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func Submit(w http.ResponseWriter, r *http.Request) {
	challengeID := r.PathValue("challengeId")

	file, header, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "no video", http.StatusBadRequest)
		return
	}
	defer file.Close()

	dir := os.Getenv("DATA_DIR")
	if dir == "" {
		dir = "./data/videos"
	}
	os.MkdirAll(dir, 0755)

	dst := filepath.Join(dir, header.Filename)
	out, _ := os.Create(dst)
	defer out.Close()

	buf := make([]byte, 32*1024*1024)
	for {
		n, err := file.Read(buf)
		if n > 0 {
			out.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}

	validation := runExtractor("validate", dst)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"challenge_id": challengeID,
		"file":         header.Filename,
		"validation":   validation,
	})
}

func runExtractor(command, videoPath string) map[string]any {
	out, err := exec.Command("python3", "../extractor/main.py", command, videoPath).Output()
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	var result map[string]any
	json.Unmarshal(out, &result)
	return result
}
