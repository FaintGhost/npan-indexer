# 架构设计 — 2C2G 性能加固

## 整体架构

```
                    ┌─────────────────────────────────────────┐
                    │           2C2G Server                   │
                    │                                         │
  Client ──────►   │  ┌─────────────────────────────────┐    │
                    │  │         npan-server              │    │
                    │  │  ┌──────────────┐                │    │
                    │  │  │ RateLimit MW │                │    │
                    │  │  └──────┬───────┘                │    │
                    │  │         ▼                         │    │
                    │  │  ┌──────────────┐                │    │
                    │  │  │ Handlers     │                │    │
                    │  │  └──────┬───────┘                │    │
                    │  │         ▼                         │    │
                    │  │  ┌──────────────────────┐        │    │
                    │  │  │ CachedQueryService   │        │    │
                    │  │  │  ┌────────────────┐  │        │    │
                    │  │  │  │ LRU + TTL      │  │        │    │
                    │  │  │  └───────┬────────┘  │        │    │
                    │  │  │          ▼           │        │    │
                    │  │  │  ┌────────────────┐  │        │    │
                    │  │  │  │ QueryService   │  │        │    │
                    │  │  │  └───────┬────────┘  │        │    │
                    │  │  └──────────┼───────────┘        │    │
                    │  │             ▼                     │    │
                    │  │  ┌──────────────────────┐        │    │
                    │  │  │ SearchActivity       │        │    │
                    │  │  │ Tracker              │◄───┐   │    │
                    │  │  └──────────────────────┘    │   │    │
                    │  │                              │   │    │
                    │  │  ┌──────────────────────┐    │   │    │
                    │  │  │ SyncManager          │────┘   │    │
                    │  │  │ (dynamic throttle)   │        │    │
                    │  │  └──────────┬───────────┘        │    │
                    │  │             │                     │    │
                    │  └─────────────┼─────────────────────┘    │
                    │               ▼                           │
                    │  ┌──────────────────────┐                │
                    │  │    Meilisearch       │                │
                    │  │  (~1 GB RAM)         │                │
                    │  └──────────────────────┘                │
                    └─────────────────────────────────────────┘
```

## 组件设计

### 1. HTTP Server 超时配置

**文件**: `cmd/server/main.go`

当前 `http.Server` 只设置了 `Addr` 和 `Handler`，需要添加超时参数。

```
http.Server{
    Addr:              cfg.ServerAddr,
    Handler:           e,
    ReadHeaderTimeout: cfg.ServerReadHeaderTimeout,  // 5s
    ReadTimeout:       cfg.ServerReadTimeout,         // 10s
    WriteTimeout:      cfg.ServerWriteTimeout,        // 30s
    IdleTimeout:       cfg.ServerIdleTimeout,          // 120s
}
```

**配置扩展** (`internal/config/config.go`):

新增字段:

| 字段 | 环境变量 | 默认值 |
|------|---------|--------|
| ServerReadHeaderTimeout | SERVER_READ_HEADER_TIMEOUT | 5s |
| ServerReadTimeout | SERVER_READ_TIMEOUT | 10s |
| ServerWriteTimeout | SERVER_WRITE_TIMEOUT | 30s |
| ServerIdleTimeout | SERVER_IDLE_TIMEOUT | 120s |

### 2. 速率限制中间件挂载

**文件**: `internal/httpx/server.go`

已有 `RateLimitMiddleware` 实现（`middleware_ratelimit.go`），只需在 `NewServer()` 中挂载:

- 全局挂载搜索端点限流: `e.Use(RateLimitMiddleware(20, 40))`
- 或分组挂载:
  - `/api/v1/app/*`: 20 rps, burst 40
  - `/api/v1/admin/*`: 5 rps, burst 10

**推荐方案**: 全局挂载，管理端点再叠加更严格限流。

```
e.Use(RateLimitMiddleware(20, 40))  // 全局

admin := e.Group("/api/v1/admin",
    APIKeyAuth(adminAPIKey),
    RateLimitMiddleware(5, 10),  // 管理端点更严格
)
```

### 3. CachedQueryService

**文件**: `internal/search/cached_query_service.go` (新建)

装饰器模式，包装 `QueryService`，提供透明缓存。

```
type Searcher interface {
    Query(params models.LocalSearchParams) (QueryResult, error)
    Ping() error
}

type CachedQueryService struct {
    inner Searcher
    cache *expirable.LRU[string, QueryResult]
}
```

**缓存键生成**: 将 `LocalSearchParams` 序列化为确定性字符串（不使用 JSON，避免 map key 顺序问题）。

```
func cacheKey(p models.LocalSearchParams) string {
    // 格式: "query|type|page|pageSize|parentID|updatedAfter|updatedBefore|includeDeleted"
    // 使用 fmt.Sprintf 生成确定性字符串
}
```

**依赖注入**:

- `Handlers` 接收 `Searcher` 接口而非具体 `*QueryService`
- `main.go` 组装: `cachedService := search.NewCachedQueryService(queryService, 256, 30*time.Second)`

### 4. SearchActivityTracker

**文件**: `internal/search/activity_tracker.go` (新建)

使用原子操作记录最近搜索活动的时间戳。

```
type SearchActivityTracker struct {
    lastActive atomic.Int64  // Unix 时间戳（秒）
    windowSec  int64         // 活跃窗口（默认 5 秒）
}

func (t *SearchActivityTracker) RecordActivity()
func (t *SearchActivityTracker) IsActive() bool
```

**集成点**:

- `CachedQueryService.Query()` 内调用 `tracker.RecordActivity()`
- `SyncManager` 持有 `tracker` 引用，在每次 API 请求前检查 `IsActive()`

### 5. 同步动态降速

**文件**: `internal/service/sync_manager.go` (修改)

**策略**: SyncManager 的 `RequestLimiter` 已使用 `rate.Limiter`。当搜索活跃时，调用 `SetLimit()` 降低速率。

```
// 在每次同步 API 请求前
if tracker.IsActive() {
    limiter.SetLimit(rate.Limit(normalRate / 2))
} else {
    limiter.SetLimit(rate.Limit(normalRate))
}
```

**参数**:

| 参数 | 正常值 | 降速值 |
|------|--------|--------|
| 同步请求速率 | 由 SyncMinTimeMS 控制 | 正常值 / 2 |

### 6. 运行时调优

**文件**: `Dockerfile`

```dockerfile
ENV GOMEMLIMIT=512MiB
ENV GOGC=100
```

使用 `ENV` 而非 `ARG`，允许运行时通过 `-e` 覆盖。

## 数据流

### 搜索请求流

```
Client → RateLimitMW → Handler → CachedQueryService
                                        │
                                 cache hit? ──yes──► 返回缓存
                                        │
                                        no
                                        │
                                        ▼
                              QueryService.Query()
                                        │
                                        ▼
                              MeiliIndex.Search()
                                        │
                                        ▼
                              tracker.RecordActivity()
                                        │
                                        ▼
                              存入缓存 → 返回结果
```

### 同步动态降速流

```
SyncManager (background goroutine)
        │
        ├── 准备发起 API 请求
        │
        ├── 检查 tracker.IsActive()
        │       │
        │    active? ──yes──► limiter.SetLimit(slow)
        │       │
        │       no ──────────► limiter.SetLimit(normal)
        │
        ├── limiter.Wait(ctx)
        │
        └── 发起 Npan API 请求
```

## 文件变更总结

| 文件 | 操作 | 说明 |
|------|------|------|
| `internal/config/config.go` | 修改 | 新增超时配置字段 |
| `cmd/server/main.go` | 修改 | 设置 http.Server 超时 |
| `internal/httpx/server.go` | 修改 | 挂载 RateLimitMiddleware |
| `internal/search/cached_query_service.go` | 新建 | LRU 缓存装饰器 |
| `internal/search/activity_tracker.go` | 新建 | 搜索活动追踪器 |
| `internal/search/query_service.go` | 修改 | 提取 Searcher 接口 |
| `internal/httpx/handlers.go` | 修改 | 使用 Searcher 接口 |
| `internal/service/sync_manager.go` | 修改 | 集成动态降速 |
| `Dockerfile` | 修改 | 添加 GOMEMLIMIT/GOGC |
| `go.mod` | 修改 | 添加 golang-lru/v2 依赖 |
