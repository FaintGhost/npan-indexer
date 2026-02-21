package service

import (
	"testing"

	"npan/internal/models"
)

func TestResolveMode_AutoNoState(t *testing.T) {
	t.Parallel()

	got := resolveMode(models.SyncModeAuto, nil)
	if got != models.SyncModeFull {
		t.Fatalf("expected %q, got %q", models.SyncModeFull, got)
	}
}

func TestResolveMode_AutoWithCursor(t *testing.T) {
	t.Parallel()

	state := &models.SyncState{LastSyncTime: 1700000000}
	got := resolveMode(models.SyncModeAuto, state)
	if got != models.SyncModeIncremental {
		t.Fatalf("expected %q, got %q", models.SyncModeIncremental, got)
	}
}

func TestResolveMode_ExplicitFull(t *testing.T) {
	t.Parallel()

	state := &models.SyncState{LastSyncTime: 1700000000}
	got := resolveMode(models.SyncModeFull, state)
	if got != models.SyncModeFull {
		t.Fatalf("expected %q, got %q", models.SyncModeFull, got)
	}
}

func TestResolveMode_ExplicitIncremental(t *testing.T) {
	t.Parallel()

	got := resolveMode(models.SyncModeIncremental, nil)
	if got != models.SyncModeIncremental {
		t.Fatalf("expected %q, got %q", models.SyncModeIncremental, got)
	}
}

func TestResolveMode_AutoZeroCursor(t *testing.T) {
	t.Parallel()

	state := &models.SyncState{LastSyncTime: 0}
	got := resolveMode(models.SyncModeAuto, state)
	if got != models.SyncModeFull {
		t.Fatalf("expected %q, got %q", models.SyncModeFull, got)
	}
}
