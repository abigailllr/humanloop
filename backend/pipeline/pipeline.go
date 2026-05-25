package pipeline

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		cmd := exec.CommandContext(ctx, "python3", args...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		cancel()

		record := func() map[string]any {
			if err != nil {
				errMsg := err.Error()
				if s := stderr.String(); s != "" {
					errMsg = s
				}
				metrics.SubmissionsTotal.WithLabelValues("failed").Inc()
				metrics.ActiveJobs.Dec()
				metrics.PipelineDuration.Observe(time.Since(start).Seconds())
				p.results.Store(job.SubmissionID, JobResult{SubmissionID: job.SubmissionID, Status: StatusFailed, Error: errMsg})
				if db.Pool != nil {
					db.UpdateSubmissionStatus(context.Background(), job.SubmissionID, "failed", "", 0)
					if count, err := db.IncrementRetryCount(context.Background(), job.SubmissionID); err == nil && count >= 3 {
						db.MarkDLQ(context.Background(), job.SubmissionID)
					}
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
		metrics.SubmissionsByRobot.WithLabelValues(robot).Inc()
		metrics.QualityScore.Observe(qs)
		metrics.ActiveJobs.Dec()
		metrics.PipelineDuration.Observe(time.Since(start).Seconds())
		p.results.Store(job.SubmissionID, JobResult{SubmissionID: job.SubmissionID, Status: StatusDone})
		if db.Pool != nil {
			db.UpdateSubmissionStatus(context.Background(), job.SubmissionID, "done", hmdfPath, creditsPerSubmission)
			db.SetSubmissionQuality(context.Background(), job.SubmissionID, qs, extractorVersion)
			db.AddCredits(context.Background(), job.UserID, creditsPerSubmission)
			db.LogCreditTransaction(context.Background(), job.UserID, job.SubmissionID, "submission", creditsPerSubmission)
			db.IncrementChallengeSubmissions(context.Background(), job.ChallengeID)
			go dispatchWebhooks(job.SubmissionID, job.ChallengeID, qs)
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

var webhookClient = &http.Client{Timeout: 10 * time.Second}

func dispatchWebhooks(submissionID, challengeID string, qs float64) {
	if db.Pool == nil {
		return
	}
	datasets, err := db.GetDatasets(context.Background())
	if err != nil {
		return
	}
	payload, _ := json.Marshal(map[string]any{
		"event":         "submission.done",
		"submission_id": submissionID,
		"challenge_id":  challengeID,
		"quality_score": qs,
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	})
	for _, dataset := range datasets {
		if dataset.ChallengeID != "" && dataset.ChallengeID != challengeID {
			continue
		}
		webhooks, err := db.GetActiveWebhooksForDataset(context.Background(), dataset.ID)
		if err != nil {
			continue
		}
		for _, wh := range webhooks {
			go sendWebhook(wh.URL, wh.SecretHash, dataset.ID, payload)
		}
	}
}

func sendWebhook(url, secretHash, datasetID string, payload []byte) {
	mac := hmac.New(sha256.New, []byte(secretHash))
	mac.Write(payload)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	for _, delay := range []time.Duration{0, 30 * time.Second, 5 * time.Minute} {
		if delay > 0 {
			time.Sleep(delay)
		}
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-HumanLoop-Signature", sig)
		req.Header.Set("X-HumanLoop-Dataset", datasetID)
		resp, err := webhookClient.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < 500 {
				return
			}
		}
	}
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
