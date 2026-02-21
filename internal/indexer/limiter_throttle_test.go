package indexer

import (
  "context"
  "sync/atomic"
  "testing"
  "time"
)

type mockActivityChecker struct {
  active atomic.Bool
}

func (m *mockActivityChecker) IsActive() bool {
  return m.active.Load()
}

func TestRequestLimiter_ThrottlesWhenActive(t *testing.T) {
  t.Parallel()

  mock := &mockActivityChecker{}
  mock.active.Store(true)

  limiter := NewRequestLimiter(1, 100)
  limiter.SetActivityChecker(mock)

  start := time.Now()

  for i := range 2 {
    if err := limiter.Schedule(context.Background(), func() error {
      return nil
    }); err != nil {
      t.Fatalf("task %d failed: %v", i, err)
    }
  }

  elapsed := time.Since(start)

  // When active, minTimeMS doubles from 100ms to 200ms.
  // Two tasks need one rate-limiter wait between them, so total >= 200ms.
  if elapsed < 200*time.Millisecond {
    t.Fatalf("expected >= 200ms with throttle active, got %v", elapsed)
  }
}

func TestRequestLimiter_RestoresWhenInactive(t *testing.T) {
  t.Parallel()

  mock := &mockActivityChecker{}
  mock.active.Store(true)

  limiter := NewRequestLimiter(1, 100)
  limiter.SetActivityChecker(mock)

  // Run one task under throttled rate to let the limiter apply the slow rate.
  if err := limiter.Schedule(context.Background(), func() error {
    return nil
  }); err != nil {
    t.Fatalf("throttled task failed: %v", err)
  }

  // Deactivate search activity so the rate should restore to the original.
  mock.active.Store(false)

  // Wait for the throttled interval to expire so the limiter accumulates tokens
  // at the restored rate before we start measuring.
  time.Sleep(250 * time.Millisecond)

  start := time.Now()

  for i := range 2 {
    if err := limiter.Schedule(context.Background(), func() error {
      return nil
    }); err != nil {
      t.Fatalf("task %d failed: %v", i, err)
    }
  }

  elapsed := time.Since(start)

  // With original rate (100ms interval), two tasks need ~100ms between them.
  // They should complete well under the throttled 200ms threshold.
  if elapsed >= 200*time.Millisecond {
    t.Fatalf("expected < 200ms after restoring rate, got %v", elapsed)
  }
}
