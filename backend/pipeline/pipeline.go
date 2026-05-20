package pipeline

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	Robot          string
	VideoHash      string
	ConsentVersion string
}

type JobResult struct {
	SubmissionID string `json:"submission_id"`
	Status       Status `json:"status"`
	Error        string `json:"error,omitempty"`
}

type Pipeline struct {
	jobs       chan Job
	results    sync.Map
	outDir     string
	hmdfUpload func(submissionID string, r io.Reader) (string, error)
}

func New(workers int, outDir string, hmdfUpload func(submissionID string, r io.Reader) (string, error)) *Pipeline {
	p := &Pipeline{
		jobs:       make(chan Job, 256),
		outDir:     outDir,
		hmdfUpload: hmdfUpload,
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
	return "../scripts/extract.py"
}

func (p *Pipeline) worker() {
	for job := range p.jobs {
		start := time.Now()
		metrics.ActiveJobs.Inc()
		p.results.Store(job.SubmissionID, JobResult{SubmissionID: job.SubmissionID, Status: StatusProcessing})

		robot := job.Robot
		if robot == "" {
			robot = "generic_bimanual"
		}

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
			robot,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		out, err := exec.CommandContext(ctx, "python3", args...).Output()
		cancel()

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
				if err := os.Remove(job.VideoPath); err != nil {
					fmt.Println("remove video:", err)
				}
			}
			continue
		}

		localPath, err := p.save(job.SubmissionID, record)
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

		hmdfPath := localPath
		if p.hmdfUpload != nil {
			f, err := os.Open(localPath)
			if err == nil {
				if s3Path, err := p.hmdfUpload(job.SubmissionID, f); err == nil {
					hmdfPath = s3Path
				}
				f.Close()
			}
		}

		if isLocalPath(job.VideoPath) {
			if err := os.Remove(job.VideoPath); err != nil {
				fmt.Println("remove video:", err)
			}
		}

		qs := qualityScore(record)
		extractorVersion, _ := record["hmdf_version"].(string)

		metrics.SubmissionsTotal.WithLabelValues("done").Inc()
		metrics.ActiveJobs.Dec()
		metrics.PipelineDuration.Observe(time.Since(start).Seconds())
		p.results.Store(job.SubmissionID, JobResult{SubmissionID: job.SubmissionID, Status: StatusDone})
		if db.Pool != nil {
			db.UpdateSubmissionStatus(context.Background(), job.SubmissionID, "done", hmdfPath, creditsPerSubmission)
			db.SetSubmissionQuality(context.Background(), job.SubmissionID, qs, extractorVersion)
			db.AddCredits(context.Background(), job.UserID, creditsPerSubmission)
			db.LogCreditTransaction(context.Background(), job.UserID, job.SubmissionID, "submission", creditsPerSubmission)
			db.IncrementChallengeSubmissions(context.Background(), job.ChallengeID)
		}
	}
}

func (p *Pipeline) save(submissionID string, data map[string]any) (string, error) {
	if err := os.MkdirAll(p.outDir, 0755); err != nil {
		return "", err
	}

	dst := filepath.Join(p.outDir, submissionID+".hmdf.json.gz")

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if err := json.NewEncoder(gz).Encode(data); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}

	if err := os.WriteFile(dst, buf.Bytes(), 0644); err != nil {
		return "", err
	}
	return dst, nil
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

func qualityScore(record map[string]any) float64 {
	frames, _ := record["frames"].([]any)
	if len(frames) == 0 {
		return 0
	}
	withMotorState := 0
	for _, f := range frames {
		fr, ok := f.(map[string]any)
		if !ok {
			continue
		}
		ms, _ := fr["motor_state"].(map[string]any)
		q, _ := ms["q"].([]any)
		if len(q) > 0 {
			withMotorState++
		}
	}
	coverage := float64(withMotorState) / float64(len(frames))
	fps, _ := record["fps"].(float64)
	if fps == 0 {
		fps = 30
	}
	fpsScore := fps / 30.0
	if fpsScore > 1 {
		fpsScore = 1
	}
	durationScore := float64(len(frames)) / 300.0
	if durationScore > 1 {
		durationScore = 1
	}
	return (coverage + fpsScore + durationScore) / 3.0
}
