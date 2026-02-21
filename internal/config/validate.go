package config

import (
  "fmt"
  "log/slog"
  "strings"
)

func (c Config) Validate() error {
  var errs []string

  if strings.TrimSpace(c.AdminAPIKey) == "" {
    errs = append(errs, "NPA_ADMIN_API_KEY 不能为空")
  } else if len(c.AdminAPIKey) < 16 {
    errs = append(errs, "NPA_ADMIN_API_KEY 长度不应少于 16 字符")
  }

  if c.MeiliHost == "" {
    errs = append(errs, "MEILI_HOST 不能为空")
  }
  if c.MeiliIndex == "" {
    errs = append(errs, "MEILI_INDEX 不能为空")
  }
  if c.BaseURL == "" {
    errs = append(errs, "NPA_BASE_URL 不能为空")
  }

  if c.SyncMaxConcurrent <= 0 || c.SyncMaxConcurrent > 20 {
    errs = append(errs, "NPA_SYNC_MAX_CONCURRENT 应在 1-20 之间")
  }
  if c.Retry.MaxRetries < 0 || c.Retry.MaxRetries > 10 {
    errs = append(errs, "NPA_MAX_RETRIES 应在 0-10 之间")
  }

  hasClientCreds := c.ClientID != "" && c.ClientSecret != "" && c.SubID > 0
  hasToken := c.Token != ""
  if c.AllowConfigAuthFallback && !hasClientCreds && !hasToken {
    errs = append(errs, "NPA_ALLOW_CONFIG_AUTH_FALLBACK=true 但未提供有效凭据")
  }

  if len(errs) > 0 {
    return fmt.Errorf("配置验证失败:\n  - %s", strings.Join(errs, "\n  - "))
  }
  return nil
}

func (c Config) LogValue() slog.Value {
  return slog.GroupValue(
    slog.String("ServerAddr", c.ServerAddr),
    slog.String("MetricsAddr", c.MetricsAddr),
    slog.String("BaseURL", c.BaseURL),
    slog.String("MeiliHost", c.MeiliHost),
    slog.String("MeiliIndex", c.MeiliIndex),
    slog.String("AdminAPIKey", "[REDACTED]"),
    slog.String("ClientSecret", "[REDACTED]"),
    slog.String("Token", "[REDACTED]"),
    slog.String("MeiliAPIKey", "[REDACTED]"),
  )
}
