package service

import (
  "strings"
  "testing"

  "npan/internal/models"
)

func TestBuildVerification_MatchingCounts(t *testing.T) {
  t.Parallel()

  stats := models.CrawlStats{
    FilesIndexed:    100,
    FoldersVisited:  20,
    FilesDiscovered: 100,
    SkippedFiles:    0,
  }
  result := buildVerification(120, stats)

  if !result.Verified {
    t.Errorf("expected Verified=true, got false")
  }
  if len(result.Warnings) != 0 {
    t.Errorf("expected no warnings, got %v", result.Warnings)
  }
  if result.MeiliDocCount != 120 {
    t.Errorf("expected MeiliDocCount=120, got %d", result.MeiliDocCount)
  }
  if result.CrawledDocCount != 120 {
    t.Errorf("expected CrawledDocCount=120, got %d", result.CrawledDocCount)
  }
  if result.DiscoveredDocCount != 120 {
    t.Errorf("expected DiscoveredDocCount=120, got %d", result.DiscoveredDocCount)
  }
}

func TestBuildVerification_MeiliFewerThanCrawled(t *testing.T) {
  t.Parallel()

  stats := models.CrawlStats{
    FilesIndexed:    100,
    FoldersVisited:  20,
    FilesDiscovered: 100,
    SkippedFiles:    0,
  }
  result := buildVerification(110, stats)

  found := false
  for _, w := range result.Warnings {
    if strings.Contains(w, "MeiliSearch") {
      found = true
      break
    }
  }
  if !found {
    t.Errorf("expected a warning mentioning MeiliSearch, got %v", result.Warnings)
  }
}

func TestBuildVerification_DiscoveredMoreThanIndexed(t *testing.T) {
  t.Parallel()

  stats := models.CrawlStats{
    FilesIndexed:    100,
    FoldersVisited:  20,
    FilesDiscovered: 105,
    SkippedFiles:    5,
  }
  result := buildVerification(120, stats)

  if result.SkippedCount != 5 {
    t.Errorf("expected SkippedCount=5, got %d", result.SkippedCount)
  }
  found := false
  for _, w := range result.Warnings {
    if strings.Contains(strings.ToLower(w), "gap") ||
      strings.Contains(strings.ToLower(w), "skip") ||
      strings.Contains(strings.ToLower(w), "discovered") ||
      strings.Contains(w, "跳过") ||
      strings.Contains(w, "已发现") {
      found = true
      break
    }
  }
  if !found {
    t.Errorf("expected a warning about discovered/skipped gap, got %v", result.Warnings)
  }
}

func TestBuildVerification_AllMatchNoWarnings(t *testing.T) {
  t.Parallel()

  stats := models.CrawlStats{
    FilesIndexed:    100,
    FoldersVisited:  20,
    FilesDiscovered: 100,
    SkippedFiles:    0,
  }
  result := buildVerification(120, stats)

	if len(result.Warnings) != 0 {
		t.Errorf("expected len(Warnings)==0, got %v", result.Warnings)
	}
}

func TestAppendRootEstimateWarnings_AddsWarningOnLargeGap(t *testing.T) {
	t.Parallel()

	estimate := int64(4152)
	progress := &models.SyncProgressState{
		Roots:     []int64{123456},
		RootNames: map[int64]string{123456: "PIXELHUE"},
		RootProgress: map[string]*models.RootSyncProgress{
			"123456": {
				RootFolderID:       123456,
				EstimatedTotalDocs: &estimate,
				Stats: models.CrawlStats{
					FilesIndexed:   393,
					FoldersVisited: 119,
				},
			},
		},
	}
	v := &models.SyncVerification{Warnings: []string{}}

	appendRootEstimateWarnings(v, progress)

	found := false
	for _, warning := range v.Warnings {
		if strings.Contains(warning, "PIXELHUE") && strings.Contains(warning, "4152") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected root estimate mismatch warning, got %v", v.Warnings)
	}
}

func TestAppendRootEstimateWarnings_IgnoresSmallGap(t *testing.T) {
	t.Parallel()

	estimate := int64(4152)
	progress := &models.SyncProgressState{
		Roots:     []int64{123456},
		RootNames: map[int64]string{123456: "PIXELHUE"},
		RootProgress: map[string]*models.RootSyncProgress{
			"123456": {
				RootFolderID:       123456,
				EstimatedTotalDocs: &estimate,
				Stats: models.CrawlStats{
					FilesIndexed:   4032,
					FoldersVisited: 120,
				},
			},
		},
	}
	v := &models.SyncVerification{Warnings: []string{}}

	appendRootEstimateWarnings(v, progress)

	if len(v.Warnings) != 0 {
		t.Fatalf("expected no warning for small gap, got %v", v.Warnings)
	}
}
