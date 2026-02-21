package config

import (
  "log/slog"
  "testing"

  "npan/internal/npan"
)

// findAttr 在 slog.Value (Group 类型) 的 Attrs 中按 key 查找值
func findAttr(t *testing.T, val slog.Value, key string) (slog.Value, bool) {
  t.Helper()
  if val.Kind() != slog.KindGroup {
    t.Fatalf("expected slog.KindGroup, got %s", val.Kind())
  }
  for _, attr := range val.Group() {
    if attr.Key == key {
      return attr.Value, true
    }
  }
  return slog.Value{}, false
}

func TestConfig_LogValue_RedactsAdminAPIKey(t *testing.T) {
  cfg := Config{
    AdminAPIKey: "real-secret-key-minimum-16",
    SubType:     npan.TokenSubjectUser,
  }

  got := cfg.LogValue()

  val, ok := findAttr(t, got, "AdminAPIKey")
  if !ok {
    t.Fatal("AdminAPIKey not found in LogValue output")
  }
  if val.String() != "[REDACTED]" {
    t.Errorf("expected AdminAPIKey to be [REDACTED], got %q", val.String())
  }
}

func TestConfig_LogValue_RedactsClientSecret(t *testing.T) {
  cfg := Config{
    ClientSecret: "super-secret-client-secret",
    SubType:      npan.TokenSubjectUser,
  }

  got := cfg.LogValue()

  val, ok := findAttr(t, got, "ClientSecret")
  if !ok {
    t.Fatal("ClientSecret not found in LogValue output")
  }
  if val.String() != "[REDACTED]" {
    t.Errorf("expected ClientSecret to be [REDACTED], got %q", val.String())
  }
}

func TestConfig_LogValue_RedactsToken(t *testing.T) {
  cfg := Config{
    Token:   "bearer-token-value",
    SubType: npan.TokenSubjectUser,
  }

  got := cfg.LogValue()

  val, ok := findAttr(t, got, "Token")
  if !ok {
    t.Fatal("Token not found in LogValue output")
  }
  if val.String() != "[REDACTED]" {
    t.Errorf("expected Token to be [REDACTED], got %q", val.String())
  }
}

func TestConfig_LogValue_RedactsMeiliAPIKey(t *testing.T) {
  cfg := Config{
    MeiliAPIKey: "meili-master-key-value",
    SubType:     npan.TokenSubjectUser,
  }

  got := cfg.LogValue()

  val, ok := findAttr(t, got, "MeiliAPIKey")
  if !ok {
    t.Fatal("MeiliAPIKey not found in LogValue output")
  }
  if val.String() != "[REDACTED]" {
    t.Errorf("expected MeiliAPIKey to be [REDACTED], got %q", val.String())
  }
}

func TestConfig_LogValue_ShowsNonSensitiveFields(t *testing.T) {
  cfg := Config{
    ServerAddr: ":1323",
    BaseURL:    "https://npan.example.com/openapi",
    MeiliHost:  "http://127.0.0.1:7700",
    SubType:    npan.TokenSubjectUser,
  }

  got := cfg.LogValue()

  cases := []struct {
    key  string
    want string
  }{
    {"ServerAddr", ":1323"},
    {"BaseURL", "https://npan.example.com/openapi"},
    {"MeiliHost", "http://127.0.0.1:7700"},
  }

  for _, tc := range cases {
    val, ok := findAttr(t, got, tc.key)
    if !ok {
      t.Errorf("field %q not found in LogValue output", tc.key)
      continue
    }
    if val.String() != tc.want {
      t.Errorf("field %q: expected %q, got %q", tc.key, tc.want, val.String())
    }
  }
}
