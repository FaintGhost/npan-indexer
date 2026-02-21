package config

import (
  "strings"
  "testing"

  "npan/internal/models"
  "npan/internal/npan"
)

// validConfig 返回一个通过所有验证的基准配置
func validConfig() Config {
  return Config{
    AdminAPIKey:       "a-valid-admin-api-key", // >= 16 chars
    BaseURL:           "https://npan.example.com/openapi",
    MeiliHost:         "http://127.0.0.1:7700",
    MeiliIndex:        "npan_items",
    SyncMaxConcurrent: 5,
    SubType:           npan.TokenSubjectUser,
    Retry: models.RetryPolicyOptions{
      MaxRetries:  3,
      BaseDelayMS: 500,
      MaxDelayMS:  5000,
      JitterMS:    200,
    },
  }
}

func TestValidate_EmptyAdminAPIKey_ReturnsError(t *testing.T) {
  cfg := validConfig()
  cfg.AdminAPIKey = ""

  err := cfg.Validate()

  if err == nil {
    t.Fatal("expected error for empty AdminAPIKey, got nil")
  }
  if !strings.Contains(err.Error(), "NPA_ADMIN_API_KEY") {
    t.Errorf("expected error message to contain %q, got: %s", "NPA_ADMIN_API_KEY", err.Error())
  }
}

func TestValidate_ShortAdminAPIKey_ReturnsError(t *testing.T) {
  cfg := validConfig()
  cfg.AdminAPIKey = "short" // < 16 chars

  err := cfg.Validate()

  if err == nil {
    t.Fatal("expected error for short AdminAPIKey, got nil")
  }
}

func TestValidate_ValidConfig_NoError(t *testing.T) {
  cfg := validConfig()

  err := cfg.Validate()

  if err != nil {
    t.Errorf("expected no error for valid config, got: %v", err)
  }
}

func TestValidate_MissingMeiliHost_ReturnsError(t *testing.T) {
  cfg := validConfig()
  cfg.MeiliHost = ""

  err := cfg.Validate()

  if err == nil {
    t.Fatal("expected error for empty MeiliHost, got nil")
  }
}

func TestValidate_MissingMeiliIndex_ReturnsError(t *testing.T) {
  cfg := validConfig()
  cfg.MeiliIndex = ""

  err := cfg.Validate()

  if err == nil {
    t.Fatal("expected error for empty MeiliIndex, got nil")
  }
}

func TestValidate_MissingBaseURL_ReturnsError(t *testing.T) {
  cfg := validConfig()
  cfg.BaseURL = ""

  err := cfg.Validate()

  if err == nil {
    t.Fatal("expected error for empty BaseURL, got nil")
  }
}

func TestValidate_InvalidSyncConcurrency_ReturnsError(t *testing.T) {
  cfg := validConfig()
  cfg.SyncMaxConcurrent = 25 // > 20

  err := cfg.Validate()

  if err == nil {
    t.Fatal("expected error for SyncMaxConcurrent=25, got nil")
  }
}

func TestValidate_FallbackWithoutCredentials_ReturnsError(t *testing.T) {
  cfg := validConfig()
  cfg.AllowConfigAuthFallback = true
  cfg.Token = ""
  cfg.ClientID = ""
  cfg.ClientSecret = ""

  err := cfg.Validate()

  if err == nil {
    t.Fatal("expected error when AllowConfigAuthFallback=true but no credentials provided, got nil")
  }
}
