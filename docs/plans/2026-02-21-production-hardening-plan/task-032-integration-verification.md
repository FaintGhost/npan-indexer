# Task 032: Integration verification

**depends-on**: task-001, task-003, task-005, task-007, task-009, task-011, task-013, task-015, task-017, task-019, task-021, task-022, task-024, task-026, task-027, task-028, task-029

## Description

Final integration verification to ensure all production hardening changes work together. Run the full test suite, verify acceptance criteria from the design document, and check for regressions.

## Execution Context

**Task Number**: 032 of 032
**Phase**: Cleanup & Verification
**Prerequisites**: All implementation tasks must be complete

## BDD Scenario Reference

**Spec**: `../2026-02-21-production-hardening-design/bdd-specs.md`
**Scenarios**: All scenarios across all 6 features

## Files to Modify/Create

- No new files — verification only

## Steps

### Step 1: Run full test suite

- `go test ./... -count=1` — all tests pass
- `go test -race ./...` — no race conditions
- `go vet ./...` — no static analysis issues

### Step 2: Verify acceptance criteria

Per design document `_index.md`:
- [ ] 无 Key 访问 `/app` 可正常使用搜索和下载
- [ ] 无 `X-API-Key` 头调用 `/api/v1/search/*` 返回 401
- [ ] API Key 为空时服务拒绝启动
- [ ] `pageSize=10000` 返回 400
- [ ] 注入 `type=file OR is_deleted = true` 返回 400
- [ ] 故意触发内部错误时，响应体不含堆栈或内部路径
- [ ] 高频请求触发 429
- [ ] `docker build` 成功，镜像 < 50MB，非 root 运行
- [ ] `GET /readyz` 在 Meili 不可达时返回 503

### Step 3: Code quality check

- Verify no `err.Error()` is returned directly to clients in any handler
- Verify no `io.ReadAll` without `io.LimitReader` in client code
- Verify `requireAPIAccess` is completely removed
- Verify `c.QueryParam("token")` is removed

### Step 4: Verify git cleanliness

- `git ls-files .env .env.meilisearch` returns empty
- `.env.example` exists
- No debug/temporary code left in the codebase

## Verification Commands

```bash
go test ./... -count=1
go test -race ./...
go vet ./...
git ls-files .env .env.meilisearch
```

## Success Criteria

- All tests pass (including race detector)
- All acceptance criteria verified
- No code quality issues
- Git repository clean of secrets
- Ready for production deployment
