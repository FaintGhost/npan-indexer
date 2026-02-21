package metrics

import "github.com/prometheus/client_golang/prometheus"

// SyncMetrics holds Prometheus metrics for sync tasks.
type SyncMetrics struct {
	TasksTotal              *prometheus.CounterVec
	DurationSeconds         *prometheus.HistogramVec
	FilesIndexedTotal       *prometheus.CounterVec
	FilesFailedTotal        *prometheus.CounterVec
	Running                 prometheus.Gauge
	IncrementalChangesTotal *prometheus.CounterVec
}

// NewSyncMetrics creates and registers sync metrics with the given registerer.
func NewSyncMetrics(reg prometheus.Registerer) *SyncMetrics {
	m := &SyncMetrics{
		TasksTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "npan_sync_tasks_total",
			Help: "Total number of completed sync tasks.",
		}, []string{"mode", "status"}),
		DurationSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "npan_sync_duration_seconds",
			Help:    "Duration of sync tasks in seconds.",
			Buckets: []float64{1, 5, 30, 60, 300, 600, 1800},
		}, []string{"mode"}),
		FilesIndexedTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "npan_sync_files_indexed_total",
			Help: "Total number of files indexed during sync.",
		}, []string{"mode"}),
		FilesFailedTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "npan_sync_files_failed_total",
			Help: "Total number of failed file index operations.",
		}, []string{"mode"}),
		Running: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "npan_sync_running",
			Help: "1 if a sync task is currently running, 0 otherwise.",
		}),
		IncrementalChangesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "npan_sync_incremental_changes_total",
			Help: "Total number of incremental sync change items by operation type.",
		}, []string{"op"}),
	}
	reg.MustRegister(
		m.TasksTotal,
		m.DurationSeconds,
		m.FilesIndexedTotal,
		m.FilesFailedTotal,
		m.Running,
		m.IncrementalChangesTotal,
	)
	return m
}
