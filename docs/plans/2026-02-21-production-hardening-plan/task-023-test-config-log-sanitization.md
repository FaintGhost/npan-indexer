# Task 023: Test config log sanitization

**depends-on**: (none)

## Description

Write tests verifying that the config struct's log output sanitizes sensitive fields. When the config is logged via slog, fields like AdminAPIKey, ClientSecret, MeiliAPIKey, and Token should display as "[REDACTED]" instead of actual values.

## Execution Context

**Task Number**: 023 of 032
**Phase**: Credential Security
**Prerequisites**: None — test task

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenario**: Feature 5 "凭据管理" — "日志不打印敏感字段"

## Files to Modify/Create

- Create: `internal/config/config_log_test.go`

## Steps

### Step 1: Verify Scenario

- Confirm scenario about log sanitization exists in Feature 5

### Step 2: Implement Tests (Red)

- Create `internal/config/config_log_test.go` with:
  - `TestConfig_LogValue_RedactsAdminAPIKey` — create config with AdminAPIKey="real-secret-key"; log it; verify output contains "[REDACTED]" NOT "real-secret-key"
  - `TestConfig_LogValue_RedactsClientSecret` — same for ClientSecret
  - `TestConfig_LogValue_RedactsMeiliAPIKey` — same for MeiliAPIKey (if field exists)
  - `TestConfig_LogValue_RedactsToken` — same for Token
  - `TestConfig_LogValue_ShowsNonSensitiveFields` — ServerAddr, BaseURL, MeiliHost should appear in log output
- Tests use `slog.Default()` with a custom handler to capture log output, or use `Config.LogValue()` directly and inspect the returned `slog.Value`
- **Verification**: Tests FAIL since `LogValue()` doesn't exist on Config

## Verification Commands

```bash
go test ./internal/config/ -run TestConfig_LogValue -v
```

## Success Criteria

- Tests verify each sensitive field is redacted
- Tests verify non-sensitive fields are visible
- Uses Go's `slog.LogValuer` interface
