package config

import (
  "testing"
  "time"
)

func TestLoad_DefaultTimeouts(t *testing.T) {
  t.Setenv("SERVER_READ_HEADER_TIMEOUT", "")
  t.Setenv("SERVER_READ_TIMEOUT", "")
  t.Setenv("SERVER_WRITE_TIMEOUT", "")
  t.Setenv("SERVER_IDLE_TIMEOUT", "")

  cfg := Load()

  cases := []struct {
    name string
    got  time.Duration
    want time.Duration
  }{
    {"ServerReadHeaderTimeout", cfg.ServerReadHeaderTimeout, 5 * time.Second},
    {"ServerReadTimeout", cfg.ServerReadTimeout, 10 * time.Second},
    {"ServerWriteTimeout", cfg.ServerWriteTimeout, 30 * time.Second},
    {"ServerIdleTimeout", cfg.ServerIdleTimeout, 120 * time.Second},
  }

  for _, tc := range cases {
    if tc.got != tc.want {
      t.Errorf("%s: expected %v, got %v", tc.name, tc.want, tc.got)
    }
  }
}

func TestLoad_CustomTimeouts(t *testing.T) {
  t.Setenv("SERVER_READ_HEADER_TIMEOUT", "8s")
  t.Setenv("SERVER_READ_TIMEOUT", "15s")
  t.Setenv("SERVER_WRITE_TIMEOUT", "60s")
  t.Setenv("SERVER_IDLE_TIMEOUT", "180s")

  cfg := Load()

  cases := []struct {
    name string
    got  time.Duration
    want time.Duration
  }{
    {"ServerReadHeaderTimeout", cfg.ServerReadHeaderTimeout, 8 * time.Second},
    {"ServerReadTimeout", cfg.ServerReadTimeout, 15 * time.Second},
    {"ServerWriteTimeout", cfg.ServerWriteTimeout, 60 * time.Second},
    {"ServerIdleTimeout", cfg.ServerIdleTimeout, 180 * time.Second},
  }

  for _, tc := range cases {
    if tc.got != tc.want {
      t.Errorf("%s: expected %v, got %v", tc.name, tc.want, tc.got)
    }
  }
}

func TestLoad_DefaultStateDBFile(t *testing.T) {
  t.Setenv("NPA_STATE_DB_FILE", "")

  cfg := Load()

  if cfg.StateDBFile != "./data/state/sync-state.sqlite" {
    t.Fatalf("expected default state db file, got %q", cfg.StateDBFile)
  }
}

func TestLoad_CustomStateDBFile(t *testing.T) {
  t.Setenv("NPA_STATE_DB_FILE", "./tmp/custom-sync-state.sqlite")

  cfg := Load()

  if cfg.StateDBFile != "./tmp/custom-sync-state.sqlite" {
    t.Fatalf("expected custom state db file, got %q", cfg.StateDBFile)
  }
}
