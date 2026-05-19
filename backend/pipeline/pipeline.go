package pipeline

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/abigailtech/humanloop/backend/db"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusDone       Status = "done"
	StatusFailed     Status = "failed"
)

const creditsPerSubmission = 10

type Job struct {
	SubmissionID   string
	ChallengeID    string
	ChallengeTitle string
	UserID         string
	VideoPath      string
	Latitude       float64
	Longitude      float64
	CapturedAt     string
}

type JobResult struct {
	SubmissionID string `json:"submission_id"`
	Status       Status `json:"status"`
	Error        string `json:"error,omitempty"`
}

type Pipeline struct {
	jobs    chan Job
	results sync.Map
	outDir  string
}

func New(workers int, outDir string) *Pipeline {
	p := &Pipeline{
		jobs:   make(chan Job, 256),
		outDir: outDir,
	}
	for i := 0; i < workers; i++ {
		go p.worker()
	}
	return p
}

func (p *Pipeline) Enqueue(job Job) {
	p.results.Store(job.SubmissionID, JobResult{SubmissionID: job.SubmissionID, Status: StatusPending})
	p.jobs <- job
}

func (p *Pipeline) Result(submissionID string) (JobResult, bool) {
	v, ok := p.results.Load(submissionID)
	if !ok {
		return JobResult{}, false
	}
	return v.(JobResult), true
}

func extractorPath() string {
	if p := os.Getenv("EXTRACTOR_PATH"); p != "" {
		return p
	}
	return "../extractor/main.py"
}

func (p *Pipeline) worker() {
	for job := range p.jobs {
		p.results.Store(job.SubmissionID, JobResult{SubmissionID: job.SubmissionID, Status: StatusProcessing})

		args := []string{
			extractorPath(), "extract",
			job.VideoPath,
			job.SubmissionID,
			job.ChallengeID,
			job.ChallengeTitle,
			job.UserID,
			fmt.Sprintf("%f", job.Latitude),
			fmt.Sprintf("%f", job.Longitude),
			job.CapturedAt,
		}

		out, err := exec.Command("python3", args...).Output()

		if err != nil {
			p.results.Store(job.SubmissionID, JobResult{
				SubmissionID: job.SubmissionID,
				Status:       StatusFailed,
				Error:        err.Error(),
			})
			if db.Pool != nil {
				db.UpdateSubmissionStatus(context.Background(), job.SubmissionID, "failed", "", 0)
			}
			continue
		}

		var record map[string]any
		if err := json.Unmarshal(out, &record); err != nil {
			p.results.Store(job.SubmissionID, JobResult{
				SubmissionID: job.SubmissionID,
				Status:       StatusFailed,
				Error:        "invalid extractor output",
			})
			if db.Pool != nil {
				db.UpdateSubmissionStatus(context.Background(), job.SubmissionID, "failed", "", 0)
			}
			continue
		}

		hmdfPath, err := p.save(job.SubmissionID, record)
		if err != nil {
			p.results.Store(job.SubmissionID, JobResult{
				SubmissionID: job.SubmissionID,
				Status:       StatusFailed,
				Error:        err.Error(),
			})
			if db.Pool != nil {
				db.UpdateSubmissionStatus(context.Background(), job.SubmissionID, "failed", "", 0)
			}
			continue
		}

		if isLocalPath(job.VideoPath) {
			os.Remove(job.VideoPath)
		}

		p.results.Store(job.SubmissionID, JobResult{SubmissionID: job.SubmissionID, Status: StatusDone})
		if db.Pool != nil {
			db.UpdateSubmissionStatus(context.Background(), job.SubmissionID, "done", hmdfPath, creditsPerSubmission)
			db.AddCredits(context.Background(), job.UserID, creditsPerSubmission)
			db.IncrementChallengeSubmissions(context.Background(), job.ChallengeID)
		}
	}
}

func (p *Pipeline) save(submissionID string, data map[string]any) (string, error) {
	if err := os.MkdirAll(p.outDir, 0755); err != nil {
		return "", err
	}

	dst := filepath.Join(p.outDir, submissionID+".hmdf.json.gz")
	f, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()

	return dst, json.NewEncoder(gz).Encode(data)
}

func isLocalPath(path string) bool {
	return !strings.HasPrefix(path, "s3://")
}
