package metrics_test

import (
	"testing"

	"npan/internal/metrics"
)

func TestNewRegistry(t *testing.T) {
	reg := metrics.NewRegistry()

	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather() returned error: %v", err)
	}

	required := map[string]bool{
		"go_goroutines":             false,
		"go_gc_duration_seconds":    false,
		"process_cpu_seconds_total": false,
	}

	for _, mf := range families {
		if _, ok := required[mf.GetName()]; ok {
			required[mf.GetName()] = true
		}
	}

	for name, found := range required {
		if !found {
			t.Errorf("expected metric family %q to be present, but it was not", name)
		}
	}
}
