# Task 005: Test CachedQueryService

**depends-on**: (none)

## Description

为 CachedQueryService（LRU + TTL 搜索缓存装饰器）创建测试。使用 mock 替代 QueryService 来隔离 Meilisearch 外部依赖。

## Execution Context

**Task Number**: 005 of 012
**Phase**: Search Cache
**Prerequisites**: None

## BDD Scenario Reference

**Spec**: `../2026-02-21-performance-hardening-design/bdd-specs.md`
**Scenario**: Scenario 3.1 (缓存命中), Scenario 3.2 (缓存过期), Scenario 3.3 (不同参数独立缓存), Scenario 3.4 (LRU 容量淘汰)

## Files to Modify/Create

- Create: `internal/search/cached_query_service_test.go`

## Steps

### Step 1: Verify Scenario

- 确认 BDD specs 中 Scenario 3.1-3.4 存在

### Step 2: Create mock

- 在测试文件中创建 `mockSearcher` 实现 `Searcher` 接口（接口尚未定义，测试先假设存在）
- `mockSearcher` 记录 `Query` 调用次数，返回预设结果

### Step 3: Implement Tests (Red)

- `TestCachedQueryService_CacheHit` (Scenario 3.1): 同参数调用两次 Query，验证 mockSearcher 只被调用一次
- `TestCachedQueryService_CacheExpiry` (Scenario 3.2): 使用极短 TTL（如 50ms），调用 Query，等待 TTL 过期后再次调用，验证 mock 被调用两次
- `TestCachedQueryService_DifferentParams` (Scenario 3.3): 用不同参数调用两次 Query，验证 mock 被调用两次
- `TestCachedQueryService_LRUEviction` (Scenario 3.4): 使用 capacity=2，插入 3 个不同参数的查询，验证第一个被淘汰（再次查询时 mock 被调用）
- **Verification**: 测试应编译失败（Red），因为 `CachedQueryService` 和 `Searcher` 接口尚未定义

### Step 4: Verify Red

- 运行测试确认编译失败

## Verification Commands

```bash
go test ./internal/search/... -run "TestCachedQueryService" -v
```

## Success Criteria

- 测试编译失败（Red），因为缺少 CachedQueryService 和 Searcher 类型
- 测试逻辑正确映射 BDD Scenario 3.1-3.4
- 使用 mock 隔离外部依赖
