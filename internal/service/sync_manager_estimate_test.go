package service

import (
	"testing"

	"npan/internal/models"
)

func TestCreateInitialProgress_IncludesEstimatedTotalDocs(t *testing.T) {
	t.Parallel()

	progress := createInitialProgress(struct {
		Roots              []int64
		RootCheckpointMap  map[int64]string
		RootEstimateMap    map[int64]int64
		RootNameMap        map[int64]string
		StartedAt          int64
		MeiliHost          string
		MeiliIndex         string
		CheckpointTemplate string
	}{
		Roots:             []int64{1001, 1002},
		RootCheckpointMap: map[int64]string{1001: "/tmp/1001.json", 1002: "/tmp/1002.json"},
		RootEstimateMap:   map[int64]int64{1001: 121},
		StartedAt:         1700000000,
		MeiliHost:         "http://127.0.0.1:7700",
		MeiliIndex:        "npan_items",
	})

	rp1 := progress.RootProgress["1001"]
	if rp1 == nil || rp1.EstimatedTotalDocs == nil || *rp1.EstimatedTotalDocs != 121 {
		t.Fatalf("expected root 1001 estimatedTotalDocs=121, got %#v", rp1)
	}

	rp2 := progress.RootProgress["1002"]
	if rp2 == nil {
		t.Fatal("missing root 1002 progress")
	}
	if rp2.EstimatedTotalDocs != nil {
		t.Fatalf("expected root 1002 no estimate, got %#v", *rp2.EstimatedTotalDocs)
	}
}

func TestRestoreProgress_RefreshesEstimatedTotalDocs(t *testing.T) {
	t.Parallel()

	oldEstimate := int64(11)
	existing := &models.SyncProgressState{
		StartedAt: 1700000000,
		Roots:     []int64{1001},
		RootProgress: map[string]*models.RootSyncProgress{
			"1001": {
				RootFolderID:       1001,
				Status:             "pending",
				EstimatedTotalDocs: &oldEstimate,
			},
		},
	}

	restored := restoreProgress(
		existing,
		[]int64{1001},
		map[int64]string{1001: "/tmp/1001.json"},
		map[int64]int64{1001: 333},
		nil,
	)

	rp := restored.RootProgress["1001"]
	if rp == nil || rp.EstimatedTotalDocs == nil || *rp.EstimatedTotalDocs != 333 {
		t.Fatalf("expected refreshed estimate=333, got %#v", rp)
	}
}
