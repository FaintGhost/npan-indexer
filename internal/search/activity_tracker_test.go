package search

import (
  "testing"
  "time"
)

func TestSearchActivityTracker_RecordAndIsActive(t *testing.T) {
  tracker := NewSearchActivityTracker(2)
  tracker.RecordActivity()

  if !tracker.IsActive() {
    t.Fatal("expected IsActive() to return true immediately after RecordActivity()")
  }
}

func TestSearchActivityTracker_ExpiresAfterWindow(t *testing.T) {
  tracker := NewSearchActivityTracker(1)
  tracker.RecordActivity()

  time.Sleep(1100 * time.Millisecond)

  if tracker.IsActive() {
    t.Fatal("expected IsActive() to return false after window expired")
  }
}

func TestSearchActivityTracker_InitiallyInactive(t *testing.T) {
  tracker := NewSearchActivityTracker(2)

  if tracker.IsActive() {
    t.Fatal("expected IsActive() to return false when no activity has been recorded")
  }
}
