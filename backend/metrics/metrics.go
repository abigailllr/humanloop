package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	SubmissionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "humanloop_submissions_total",
	}, []string{"status"})

	SubmissionsByRobot = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "humanloop_submissions_by_robot_total",
	}, []string{"robot"})

	PipelineDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "humanloop_pipeline_duration_seconds",
		Buckets: prometheus.DefBuckets,
	})

	QualityScore = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "humanloop_submission_quality_score",
		Buckets: []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
	})

	ActiveJobs = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "humanloop_pipeline_active_jobs",
	})

	QueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "humanloop_queue_depth",
	})
)
