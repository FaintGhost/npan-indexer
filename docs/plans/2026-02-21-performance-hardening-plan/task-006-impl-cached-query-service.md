# Task 006: Implement CachedQueryService

**depends-on**: task-005

## Description

创建 CachedQueryService，使用 `hashicorp/golang-lru/v2/expirable` 实现带 TTL 的 LRU 缓存装饰器。同时定义 `Searcher` 接口，修改 `NewHandlers` 接受接口而非具体类型。

## Execution Context

**Task Number**: 006 of 012
**Phase**: Search Cache
**Prerequisites**: Task 005 测试已创建

## BDD Scenario Reference

**Spec**: `../2026-02-21-performance-hardening-design/bdd-specs.md`
**Scenario**: Scenario 3.1 (缓存命中), Scenario 3.2 (缓存过期), Scenario 3.3 (不同参数独立缓存), Scenario 3.4 (LRU 容量淘汰)

## Files to Modify/Create

- Create: `internal/search/cached_query_service.go` — CachedQueryService 实现
- Modify: `internal/search/query_service.go` — 提取 Searcher 接口
- Modify: `internal/httpx/handlers.go` — 修改 `NewHandlers` 签名接受 `search.Searcher` 接口
- Modify: `go.mod` — 添加 `github.com/hashicorp/golang-lru/v2` 依赖

## Steps

### Step 1: Add dependency

- 运行 `go get github.com/hashicorp/golang-lru/v2`

### Step 2: Define Searcher interface

- 在 `query_service.go` 中定义 `Searcher` 接口，包含 `Query(models.LocalSearchParams) (QueryResult, error)` 和 `Ping() error` 方法
- `QueryService` 已隐式实现此接口

### Step 3: Create CachedQueryService

- 在 `cached_query_service.go` 中创建 `CachedQueryService` 结构体
- 内嵌 `inner Searcher` 和 `cache *expirable.LRU[string, QueryResult]`
- 构造函数 `NewCachedQueryService(inner Searcher, capacity int, ttl time.Duration) *CachedQueryService`
- `cacheKey` 函数：将 `LocalSearchParams` 序列化为确定性字符串（使用 `strings.Builder` + `fmt.Fprintf`，不用 JSON）
- `Query` 方法：先查缓存，命中直接返回；未命中调用 `inner.Query()`，结果存入缓存后返回
- `Ping` 方法：直接委托给 `inner.Ping()`

### Step 4: Update NewHandlers signature

- 修改 `handlers.go` 中 `NewHandlers` 的 `queryService` 参数类型从 `*search.QueryService` 改为 `search.Searcher`

### Step 5: Verify Green

- 运行 Task 005 创建的测试，验证全部通过
- **Verification**: `go test ./internal/search/... -run "TestCachedQueryService" -v`

## Verification Commands

```bash
go get github.com/hashicorp/golang-lru/v2
go test ./internal/search/... -run "TestCachedQueryService" -v
go build ./...
```

## Success Criteria

- Task 005 的测试全部通过（Green）
- 全项目编译成功
- `go.mod` 包含 `hashicorp/golang-lru/v2`
