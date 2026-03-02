package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	ActiveSessions = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "exam_active_sessions",
			Help: "Number of active exam sessions",
		},
	)

	SyncLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "exam_sync_latency_seconds",
			Help:    "Latency of sync endpoint",
			Buckets: prometheus.DefBuckets,
		},
	)

	SyncFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "exam_sync_failures_total",
			Help: "Total sync failures",
		},
	)

	GraceUsage = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "exam_grace_usage_total",
			Help: "Total grace usage count",
		},
	)

)

func Init() {
	prometheus.MustRegister(
		ActiveSessions,
		SyncLatency,
		SyncFailures,
		GraceUsage,
	)
}