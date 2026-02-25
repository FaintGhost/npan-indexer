package service

import (
	"testing"

	"npan/internal/models"
)

func TestResolveMode_EmptyDefaultsToFull(t *testing.T) {
	t.Parallel()

	got, err := resolveMode("")
	if err != nil {
		t.Fatalf("resolveMode returned error: %v", err)
	}
	if got != models.SyncModeFull {
		t.Fatalf("expected %q, got %q", models.SyncModeFull, got)
	}
}

func TestResolveMode_ExplicitFull(t *testing.T) {
	t.Parallel()

	got, err := resolveMode(models.SyncModeFull)
	if err != nil {
		t.Fatalf("resolveMode returned error: %v", err)
	}
	if got != models.SyncModeFull {
		t.Fatalf("expected %q, got %q", models.SyncModeFull, got)
	}
}

func TestResolveMode_ExplicitIncremental(t *testing.T) {
	t.Parallel()

	got, err := resolveMode(models.SyncModeIncremental)
	if err != nil {
		t.Fatalf("resolveMode returned error: %v", err)
	}
	if got != models.SyncModeIncremental {
		t.Fatalf("expected %q, got %q", models.SyncModeIncremental, got)
	}
}

func TestResolveMode_InvalidModeReturnsError(t *testing.T) {
	t.Parallel()

	if _, err := resolveMode(models.SyncMode("auto")); err == nil {
		t.Fatalf("expected error for invalid mode")
	}
}
