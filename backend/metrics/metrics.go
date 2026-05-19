package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	SubmissionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "humanloop_submissions_total",
		Help: "Total submissions by outcome",
	}, []string{"status"})

	PipelineDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "humanloop_pipeline_duration_seconds",
		Help:    "Time to process a submission through the full pipeline",
		Buckets: prometheus.DefBuckets,
	})

	ActiveJobs = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "humanloop_pipeline_active_jobs",
		Help: "Number of jobs currently being processed",
	})
)
