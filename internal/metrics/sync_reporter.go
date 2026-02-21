package metrics

import (
	"time"

	"npan/internal/models"
)

// SyncEvent describes a completed sync task event.
type SyncEvent struct {
	Mode      models.SyncMode
	Status    string // "done" | "error" | "cancelled"
	Duration  time.Duration
	Stats     models.CrawlStats
	IncrStats *models.IncrementalSyncStats
}

// SyncReporter reports sync lifecycle events to metrics. Can be nil (disabled).
type SyncReporter interface {
	ReportSyncStarted(mode models.SyncMode)
	ReportSyncFinished(event SyncEvent)
}

// PrometheusSyncReporter implements SyncReporter using SyncMetrics.
type PrometheusSyncReporter struct {
	m *SyncMetrics
}

// NewPrometheusSyncReporter creates a new reporter backed by the given SyncMetrics.
func NewPrometheusSyncReporter(m *SyncMetrics) *PrometheusSyncReporter {
	return &PrometheusSyncReporter{m: m}
}

func (r *PrometheusSyncReporter) ReportSyncStarted(mode models.SyncMode) {
	r.m.Running.Set(1)
}

func (r *PrometheusSyncReporter) ReportSyncFinished(event SyncEvent) {
	modeStr := string(event.Mode)
	r.m.Running.Set(0)
	r.m.TasksTotal.WithLabelValues(modeStr, event.Status).Inc()
	r.m.DurationSeconds.WithLabelValues(modeStr).Observe(event.Duration.Seconds())
	r.m.FilesIndexedTotal.WithLabelValues(modeStr).Add(float64(event.Stats.FilesIndexed))
	r.m.FilesFailedTotal.WithLabelValues(modeStr).Add(float64(event.Stats.FailedRequests))

	if event.IncrStats != nil {
		r.m.IncrementalChangesTotal.WithLabelValues("upsert").Add(float64(event.IncrStats.Upserted))
		r.m.IncrementalChangesTotal.WithLabelValues("delete").Add(float64(event.IncrStats.Deleted))
		r.m.IncrementalChangesTotal.WithLabelValues("skip_upsert").Add(float64(event.IncrStats.SkippedUpserts))
		r.m.IncrementalChangesTotal.WithLabelValues("skip_delete").Add(float64(event.IncrStats.SkippedDeletes))
	}
}
