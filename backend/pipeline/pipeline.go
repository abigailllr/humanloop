package pipeline

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusDone       Status = "done"
	StatusFailed     Status = "failed"
)

type Job struct {
	SubmissionID   string
	ChallengeID    string
	ChallengeTitle string
	UserID         string
	VideoPath      string
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

func (p *Pipeline) worker() {
	for job := range p.jobs {
		p.results.Store(job.SubmissionID, JobResult{SubmissionID: job.SubmissionID, Status: StatusProcessing})

		out, err := exec.Command(
			"python3", "../extractor/main.py", "extract",
			job.VideoPath,
			job.SubmissionID,
			job.ChallengeID,
			job.ChallengeTitle,
			job.UserID,
		).Output()

		if err != nil {
			p.results.Store(job.SubmissionID, JobResult{
				SubmissionID: job.SubmissionID,
				Status:       StatusFailed,
				Error:        err.Error(),
			})
			continue
		}

		var record map[string]any
		if err := json.Unmarshal(out, &record); err != nil {
			p.results.Store(job.SubmissionID, JobResult{
				SubmissionID: job.SubmissionID,
				Status:       StatusFailed,
				Error:        "invalid extractor output",
			})
			continue
		}

		if err := p.save(job.SubmissionID, record); err != nil {
			p.results.Store(job.SubmissionID, JobResult{
				SubmissionID: job.SubmissionID,
				Status:       StatusFailed,
				Error:        err.Error(),
			})
			continue
		}

		if isLocalPath(job.VideoPath) {
			os.Remove(job.VideoPath)
		}

		p.results.Store(job.SubmissionID, JobResult{SubmissionID: job.SubmissionID, Status: StatusDone})
	}
}

func (p *Pipeline) save(submissionID string, data map[string]any) error {
	if err := os.MkdirAll(p.outDir, 0755); err != nil {
		return err
	}

	dst := filepath.Join(p.outDir, submissionID+".hmdf.json.gz")
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()

	return json.NewEncoder(gz).Encode(data)
}

func isLocalPath(path string) bool {
	return !strings.HasPrefix(path, "s3://")
}
