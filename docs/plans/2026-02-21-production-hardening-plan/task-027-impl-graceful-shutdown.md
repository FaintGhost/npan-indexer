# Task 027: Implement graceful shutdown

**depends-on**: (none)

## Description

Implement graceful shutdown in `cmd/server/main.go` using `signal.NotifyContext`. When SIGINT or SIGTERM is received, the server should stop accepting new connections, finish in-flight requests, and shut down cleanly within a timeout.

## Execution Context

**Task Number**: 027 of 032
**Phase**: Health & Ops
**Prerequisites**: None

## BDD Scenario Reference

**Spec**: No direct BDD scenario
**Architecture ref**: `../2026-02-21-production-hardening-design/architecture.md` Section 8

## Files to Modify/Create

- Modify: `cmd/server/main.go` — replace direct `server.Start()` with signal-aware shutdown

## Steps

### Step 1: Implement signal handling

- In `cmd/server/main.go`, after server initialization:
  - Create `signal.NotifyContext` for SIGINT and SIGTERM
  - Start the server in a goroutine
  - Wait for context cancellation (signal received)
  - Create a shutdown timeout context (15 seconds)
  - Call `server.Shutdown(shutdownCtx)` for graceful shutdown
  - Log shutdown progress

### Step 2: Write test

- Create `cmd/server/main_test.go` with a basic test that verifies the server starts and can be shut down:
  - `TestServer_GracefulShutdown` — start server in goroutine, send SIGINT, verify it stops within timeout
  - This may need to be a simple integration test

### Step 3: Verify

- Run the test
- Manually test: start server, send SIGTERM, verify clean exit

## Verification Commands

```bash
go build ./cmd/server/
go test ./cmd/server/ -v
go vet ./cmd/server/
```

## Success Criteria

- Server exits cleanly on SIGINT/SIGTERM
- In-flight requests are completed before shutdown
- Shutdown timeout prevents hanging
- No direct `server.Start()` call blocking main goroutine
