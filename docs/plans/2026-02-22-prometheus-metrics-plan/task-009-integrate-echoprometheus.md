# Task 009: Integrate echoprometheus middleware

**depends-on**: task-001

## Description

Add `echoprometheus` middleware to the Echo server in `internal/httpx/server.go` for automatic HTTP request metrics collection. Configure Skipper, LabelFuncs, and custom histogram buckets.

## Execution Context

**Task Number**: 009 of 012
**Phase**: Integration
**Prerequisites**: echo-contrib dependency available

## BDD Scenario Reference

**Spec**: `../2026-02-22-prometheus-metrics-design/bdd-specs.md`
**Scenarios**: "成功的 API 请求被计入请求总数", "请求延迟被记录到 Histogram", "健康检查路由不被统计", "就绪检查路由不被统计", "SPA 路由被规范化"

## Files to Modify/Create

- Modify: `internal/httpx/server.go`
- Create: `internal/httpx/server_test.go` (or add to existing)

## Steps

### Step 1: Write test (Red)

Create `internal/httpx/server_test.go`. Use a fresh `prometheus.NewRegistry()`, create a minimal Echo server via `NewServer` with mock handlers, and test:

1. **Requests counted** — Send GET to `/api/v1/search/local` (mock handler returns 200). Gather metrics from the registry, verify `npan_requests_total` has an entry with `method="GET"`, `code="200"`
2. **Healthz skipped** — Send GET to `/healthz`. Gather metrics, verify no entry with `url="/healthz"` in `npan_requests_total`
3. **Readyz skipped** — Same for `/readyz`
4. **SPA normalized** — Send GET to `/settings` (falls through to SPA catch-all). Verify the `url` label is `/spa`, not `/*` or `/settings`
5. **Duration recorded** — Verify `npan_request_duration_seconds` histogram has at least one observation after a request

Note: The test needs to create mock Handlers (with stubbed methods). `NewServer` signature will change to accept `prometheus.Registerer`.

**Verification**: `go test ./internal/httpx/... -run TestPrometheusMiddleware` → MUST FAIL

### Step 2: Implement (Green)

Modify `internal/httpx/server.go`:

1. Change `NewServer` signature to add `promReg prometheus.Registerer` parameter:
   ```
   func NewServer(handlers *Handlers, adminAPIKey string, distFS fs.FS, promReg prometheus.Registerer) *echo.Echo
   ```

2. After creating `echo.New()` and before existing middleware, add `echoprometheus.NewMiddlewareWithConfig`:
   - `Subsystem`: `"npan"`
   - `Registerer`: the `promReg` parameter
   - `Skipper`: skip `/healthz` and `/readyz`
   - `LabelFuncs`: override `"url"` to normalize SPA catch-all `/*` to `/spa`
   - `HistogramOptsFunc`: for `request_duration_seconds`, use custom buckets `{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0}`

3. Add necessary imports for `echoprometheus` and `prometheus`

**Verification**: `go test ./internal/httpx/... -run TestPrometheusMiddleware` → MUST PASS

### Step 3: Fix compilation

Update all callers of `NewServer` (currently only `cmd/server/main.go`) to pass a `prometheus.Registerer`. For now, pass `prometheus.NewRegistry()` as a temporary placeholder — this will be replaced in Task 011.

**Verification**: `go build ./...` compiles

## Verification Commands

```bash
go test ./internal/httpx/... -run TestPrometheusMiddleware -v
go build ./...
```

## Success Criteria

- HTTP requests are counted with correct labels
- Health check routes are excluded
- SPA routes are normalized to `/spa`
- Custom histogram buckets are applied
- All existing code compiles
