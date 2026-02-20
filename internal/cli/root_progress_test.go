package cli

import (
	"strings"
	"testing"

	"npan/internal/models"
)

func TestResolveSyncProgressOutputMode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input   string
		want    syncProgressOutputMode
		wantErr bool
	}{
		{input: "", want: syncProgressOutputHuman},
		{input: "human", want: syncProgressOutputHuman},
		{input: "HUMAN", want: syncProgressOutputHuman},
		{input: "json", want: syncProgressOutputJSON},
		{input: "JSON", want: syncProgressOutputJSON},
		{input: "foo", wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			got, err := resolveSyncProgressOutputMode(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tc.input, err)
			}
			if got != tc.want {
				t.Fatalf("expected mode %q, got %q", tc.want, got)
			}
		})
	}
}

func TestRenderSyncFullProgressHuman(t *testing.T) {
	t.Parallel()

	activeRootID := int64(123)
	currentFolderID := int64(456)
	currentPageID := int64(2)
	currentPageCount := int64(9)
	queueLength := int64(7)
	estimatedTotalDocs := int64(120)

	progress := &models.SyncProgressState{
		Status:         "running",
		StartedAt:      1000,
		UpdatedAt:      3000,
		Roots:          []int64{123, 789},
		CompletedRoots: []int64{789},
		ActiveRoot:     &activeRootID,
		AggregateStats: models.CrawlStats{
			FilesIndexed:   30,
			PagesFetched:   9,
			FoldersVisited: 5,
			FailedRequests: 1,
		},
		RootProgress: map[string]*models.RootSyncProgress{
			"123": {
				CurrentFolderID:  &currentFolderID,
				CurrentPageID:    &currentPageID,
				CurrentPageCount: &currentPageCount,
				QueueLength:      &queueLength,
				Stats: models.CrawlStats{
					FilesIndexed:   30,
					FoldersVisited: 5,
				},
				EstimatedTotalDocs: &estimatedTotalDocs,
			},
		},
	}

	snapshot := &progressRenderSnapshot{
		updatedAtMillis: 1000,
		filesIndexed:    10,
		pagesFetched:    5,
	}

	line := renderSyncFullProgressHuman(progress, snapshot)
	if !strings.Contains(line, "status=running") {
		t.Fatalf("expected status in line, got: %s", line)
	}
	if !strings.Contains(line, "roots=1/2") {
		t.Fatalf("expected roots in line, got: %s", line)
	}
	if !strings.Contains(line, "active=123") {
		t.Fatalf("expected active root in line, got: %s", line)
	}
	if !strings.Contains(line, "files=30") || !strings.Contains(line, "pages=9") {
		t.Fatalf("expected stats in line, got: %s", line)
	}
	if !strings.Contains(line, "file_rate=10/s") {
		t.Fatalf("expected file rate in line, got: %s", line)
	}
	if !strings.Contains(line, "page_rate=2/s") {
		t.Fatalf("expected page rate in line, got: %s", line)
	}
	if !strings.Contains(line, "root{folder=456 page=3/9 queue=7}") {
		t.Fatalf("expected active root detail in line, got: %s", line)
	}
	if !strings.Contains(line, "est=29.2%") {
		t.Fatalf("expected estimate percentage in line, got: %s", line)
	}
	if !strings.Contains(line, "docs=35/120") {
		t.Fatalf("expected estimate docs in line, got: %s", line)
	}
	if !strings.Contains(line, "roots=1/2") {
		t.Fatalf("expected estimate coverage roots in line, got: %s", line)
	}

	if snapshot.updatedAtMillis != 3000 || snapshot.filesIndexed != 30 || snapshot.pagesFetched != 9 {
		t.Fatalf("snapshot not updated correctly: %#v", snapshot)
	}
}

func TestRenderSyncFullProgressHuman_EstimateNA(t *testing.T) {
	t.Parallel()

	progress := &models.SyncProgressState{
		Status:    "running",
		StartedAt: 1000,
		UpdatedAt: 2000,
		Roots:     []int64{1},
		RootProgress: map[string]*models.RootSyncProgress{
			"1": {
				Stats: models.CrawlStats{
					FilesIndexed:   3,
					FoldersVisited: 2,
				},
			},
		},
	}

	line := renderSyncFullProgressHuman(progress, &progressRenderSnapshot{})
	if !strings.Contains(line, "est=n/a") {
		t.Fatalf("expected est=n/a, got: %s", line)
	}
}
