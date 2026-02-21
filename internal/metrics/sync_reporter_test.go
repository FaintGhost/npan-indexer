package metrics_test

import (
	"testing"
	"time"

	"npan/internal/metrics"
	"npan/internal/models"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestPrometheusSyncReporter_Started(t *testing.T) {
	reg := prometheus.NewRegistry()
	sm := metrics.NewSyncMetrics(reg)
	r := metrics.NewPrometheusSyncReporter(sm)

	r.ReportSyncStarted(models.SyncModeFull)
	if v := testutil.ToFloat64(sm.Running); v != 1 {
		t.Errorf("Running after start: got %f, want 1", v)
	}
}

func TestPrometheusSyncReporter_FinishedDone(t *testing.T) {
	reg := prometheus.NewRegistry()
	sm := metrics.NewSyncMetrics(reg)
	r := metrics.NewPrometheusSyncReporter(sm)

	r.ReportSyncStarted(models.SyncModeFull)
	r.ReportSyncFinished(metrics.SyncEvent{
		Mode:     models.SyncModeFull,
		Status:   "done",
		Duration: 42 * time.Second,
		Stats:    models.CrawlStats{FilesIndexed: 500, FailedRequests: 2},
	})

	if v := testutil.ToFloat64(sm.Running); v != 0 {
		t.Errorf("Running after finish: got %f, want 0", v)
	}
	if v := testutil.ToFloat64(sm.TasksTotal.WithLabelValues("full", "done")); v != 1 {
		t.Errorf("TasksTotal full/done: got %f, want 1", v)
	}
	if v := testutil.ToFloat64(sm.FilesIndexedTotal.WithLabelValues("full")); v != 500 {
		t.Errorf("FilesIndexedTotal: got %f, want 500", v)
	}
	if v := testutil.ToFloat64(sm.FilesFailedTotal.WithLabelValues("full")); v != 2 {
		t.Errorf("FilesFailedTotal: got %f, want 2", v)
	}
}

func TestPrometheusSyncReporter_FinishedCancelled(t *testing.T) {
	reg := prometheus.NewRegistry()
	sm := metrics.NewSyncMetrics(reg)
	r := metrics.NewPrometheusSyncReporter(sm)

	r.ReportSyncFinished(metrics.SyncEvent{
		Mode:   models.SyncModeFull,
		Status: "cancelled",
	})

	if v := testutil.ToFloat64(sm.TasksTotal.WithLabelValues("full", "cancelled")); v != 1 {
		t.Errorf("TasksTotal full/cancelled: got %f, want 1", v)
	}
}

func TestPrometheusSyncReporter_FinishedError(t *testing.T) {
	reg := prometheus.NewRegistry()
	sm := metrics.NewSyncMetrics(reg)
	r := metrics.NewPrometheusSyncReporter(sm)

	r.ReportSyncFinished(metrics.SyncEvent{
		Mode:   models.SyncModeIncremental,
		Status: "error",
	})

	if v := testutil.ToFloat64(sm.TasksTotal.WithLabelValues("incremental", "error")); v != 1 {
		t.Errorf("TasksTotal incremental/error: got %f, want 1", v)
	}
}

func TestPrometheusSyncReporter_IncrementalStats(t *testing.T) {
	reg := prometheus.NewRegistry()
	sm := metrics.NewSyncMetrics(reg)
	r := metrics.NewPrometheusSyncReporter(sm)

	r.ReportSyncFinished(metrics.SyncEvent{
		Mode:   models.SyncModeIncremental,
		Status: "done",
		IncrStats: &models.IncrementalSyncStats{
			Upserted:       10,
			Deleted:        3,
			SkippedUpserts: 1,
			SkippedDeletes: 2,
		},
	})

	if v := testutil.ToFloat64(sm.IncrementalChangesTotal.WithLabelValues("upsert")); v != 10 {
		t.Errorf("incr upsert: got %f, want 10", v)
	}
	if v := testutil.ToFloat64(sm.IncrementalChangesTotal.WithLabelValues("delete")); v != 3 {
		t.Errorf("incr delete: got %f, want 3", v)
	}
	if v := testutil.ToFloat64(sm.IncrementalChangesTotal.WithLabelValues("skip_upsert")); v != 1 {
		t.Errorf("incr skip_upsert: got %f, want 1", v)
	}
	if v := testutil.ToFloat64(sm.IncrementalChangesTotal.WithLabelValues("skip_delete")); v != 2 {
		t.Errorf("incr skip_delete: got %f, want 2", v)
	}
}
