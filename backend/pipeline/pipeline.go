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
	"time"

	"github.com/abigailtech/humanloop/backend/db"
	"github.com/abigailtech/humanloop/backend/metrics"
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
		start := time.Now()
		metrics.ActiveJobs.Inc()
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

		record := func() map[string]any {
			if err != nil {
				metrics.SubmissionsTotal.WithLabelValues("failed").Inc()
				metrics.ActiveJobs.Dec()
				metrics.PipelineDuration.Observe(time.Since(start).Seconds())
				p.results.Store(job.SubmissionID, JobResult{SubmissionID: job.SubmissionID, Status: StatusFailed, Error: err.Error()})
				if db.Pool != nil {
					db.UpdateSubmissionStatus(context.Background(), job.SubmissionID, "failed", "", 0)
				}
				return nil
			}
			var r map[string]any
			if err := json.Unmarshal(out, &r); err != nil {
				metrics.SubmissionsTotal.WithLabelValues("failed").Inc()
				metrics.ActiveJobs.Dec()
				metrics.PipelineDuration.Observe(time.Since(start).Seconds())
				p.results.Store(job.SubmissionID, JobResult{SubmissionID: job.SubmissionID, Status: StatusFailed, Error: "invalid extractor output"})
				if db.Pool != nil {
					db.UpdateSubmissionStatus(context.Background(), job.SubmissionID, "failed", "", 0)
				}
				return nil
			}
			return r
		}()
		if record == nil {
			continue
		}

		if isSynthetic(record) {
			metrics.SubmissionsTotal.WithLabelValues("synthetic").Inc()
			metrics.ActiveJobs.Dec()
			metrics.PipelineDuration.Observe(time.Since(start).Seconds())
			p.results.Store(job.SubmissionID, JobResult{SubmissionID: job.SubmissionID, Status: StatusFailed, Error: "synthetic video detected"})
			if db.Pool != nil {
				db.UpdateSubmissionStatus(context.Background(), job.SubmissionID, "synthetic", "", 0)
			}
			if isLocalPath(job.VideoPath) {
				os.Remove(job.VideoPath)
			}
			continue
		}

		hmdfPath, err := p.save(job.SubmissionID, record)
		if err != nil {
			metrics.SubmissionsTotal.WithLabelValues("failed").Inc()
			metrics.ActiveJobs.Dec()
			metrics.PipelineDuration.Observe(time.Since(start).Seconds())
			p.results.Store(job.SubmissionID, JobResult{SubmissionID: job.SubmissionID, Status: StatusFailed, Error: err.Error()})
			if db.Pool != nil {
				db.UpdateSubmissionStatus(context.Background(), job.SubmissionID, "failed", "", 0)
			}
			continue
		}

		if isLocalPath(job.VideoPath) {
			os.Remove(job.VideoPath)
		}

		metrics.SubmissionsTotal.WithLabelValues("done").Inc()
		metrics.ActiveJobs.Dec()
		metrics.PipelineDuration.Observe(time.Since(start).Seconds())
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

func isSynthetic(record map[string]any) bool {
	sd, ok := record["synthetic_detection"].(map[string]any)
	if !ok {
		return false
	}
	synthetic, _ := sd["synthetic"].(bool)
	return synthetic
}
