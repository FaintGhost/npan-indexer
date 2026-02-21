package metrics_test

import (
	"testing"

	"npan/internal/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNewSyncMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := metrics.NewSyncMetrics(reg)

	// Running gauge
	m.Running.Set(1)
	if v := testutil.ToFloat64(m.Running); v != 1 {
		t.Errorf("Running: got %f, want 1", v)
	}

	// TasksTotal counter vec
	m.TasksTotal.WithLabelValues("full", "done").Inc()
	if v := testutil.ToFloat64(m.TasksTotal.WithLabelValues("full", "done")); v != 1 {
		t.Errorf("TasksTotal: got %f, want 1", v)
	}

	// DurationSeconds histogram
	m.DurationSeconds.WithLabelValues("full").Observe(42.5)

	// FilesIndexedTotal
	m.FilesIndexedTotal.WithLabelValues("full").Add(500)
	if v := testutil.ToFloat64(m.FilesIndexedTotal.WithLabelValues("full")); v != 500 {
		t.Errorf("FilesIndexedTotal: got %f, want 500", v)
	}

	// FilesFailedTotal
	m.FilesFailedTotal.WithLabelValues("incremental").Add(3)
	if v := testutil.ToFloat64(m.FilesFailedTotal.WithLabelValues("incremental")); v != 3 {
		t.Errorf("FilesFailedTotal: got %f, want 3", v)
	}

	// IncrementalChangesTotal
	m.IncrementalChangesTotal.WithLabelValues("upsert").Add(10)
	if v := testutil.ToFloat64(m.IncrementalChangesTotal.WithLabelValues("upsert")); v != 10 {
		t.Errorf("IncrementalChangesTotal: got %f, want 10", v)
	}
}
