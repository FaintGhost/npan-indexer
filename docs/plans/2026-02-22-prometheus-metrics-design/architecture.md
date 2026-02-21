# Architecture

## 包结构

```
internal/metrics/
├── registry.go            # NewRegistry() → *prometheus.Registry
├── sync_metrics.go        # SyncMetrics struct + NewSyncMetrics()
├── search_metrics.go      # SearchMetrics struct + NewSearchMetrics()
├── sync_reporter.go       # SyncReporter interface + PrometheusSyncReporter
├── search_instrumenter.go # InstrumentedSearchService (CachedQueryService 装饰器)
├── meili_instrumenter.go  # InstrumentedMeiliIndex (MeiliIndex 装饰器)
└── server.go              # StartMetricsServer() / metrics HTTP server
```

## 依赖方向

```
cmd/server/main.go
  ├─→ internal/metrics     (构建指标系统)
  ├─→ internal/httpx       (传入 promReg)
  ├─→ internal/service     (传入 SyncReporter)
  └─→ internal/search      (使用 InstrumentedMeiliIndex)

internal/metrics
  ├─→ internal/models      (SyncMode, CrawlStats 等类型)
  └─→ internal/search      (IndexOperator 接口, Searcher 接口)

internal/service
  └─→ internal/metrics     (SyncReporter 接口)

internal/httpx
  └─→ prometheus           (Registerer 接口)

无循环依赖。
```

## 组件详情

### 1. registry.go

```go
package metrics

func NewRegistry() *prometheus.Registry {
  reg := prometheus.NewRegistry()
  reg.MustRegister(
    collectors.NewGoCollector(),
    collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
  )
  return reg
}
```

### 2. sync_metrics.go

```go
type SyncMetrics struct {
  TasksTotal              *prometheus.CounterVec   // {mode, status}
  DurationSeconds         *prometheus.HistogramVec // {mode}
  FilesIndexedTotal       *prometheus.CounterVec   // {mode}
  FilesFailedTotal        *prometheus.CounterVec   // {mode}
  Running                 prometheus.Gauge
  IncrementalChangesTotal *prometheus.CounterVec   // {op}
}
```

Histogram buckets for DurationSeconds: `{1, 5, 30, 60, 300, 600, 1800}`（同步任务从秒级到半小时级）

### 3. search_metrics.go

```go
type SearchMetrics struct {
  QueriesTotal           *prometheus.CounterVec   // {result: hit/miss}
  CacheSize              prometheus.Gauge
  MeiliDurationSeconds   *prometheus.HistogramVec // {op: search/upsert/delete/document_count}
  MeiliErrorsTotal       *prometheus.CounterVec   // {op}
  MeiliDocumentsTotal    prometheus.Gauge
  MeiliUpsertedDocsTotal prometheus.Counter
}
```

Histogram buckets for MeiliDurationSeconds: `{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5}`（毫秒到秒级）

### 4. sync_reporter.go — SyncManager 的指标上报接口

```go
// SyncReporter 由 SyncManager 调用，可为 nil（禁用指标）。
type SyncReporter interface {
  ReportSyncStarted(mode models.SyncMode)
  ReportSyncFinished(event SyncEvent)
}

type SyncEvent struct {
  Mode      models.SyncMode
  Status    string // "done" | "error" | "cancelled"
  Duration  time.Duration
  Stats     models.CrawlStats
  IncrStats *models.IncrementalSyncStats
}
```

PrometheusSyncReporter 在 `ReportSyncStarted` 时设 `Running=1`，在 `ReportSyncFinished` 时更新所有计数器/直方图并重设 `Running=0`。

### 5. search_instrumenter.go — CachedQueryService 装饰器

包装 `search.Searcher` 接口，在每次 `Query()` 调用后：
- 判断缓存命中/未命中，递增 `QueriesTotal`
- 更新 `CacheSize` gauge

需要 `CachedQueryService` 暴露一个 `Len() int` 方法来获取当前缓存大小。

### 6. meili_instrumenter.go — MeiliIndex 装饰器

包装 `search.IndexOperator` 接口，为每个操作记录：
- `MeiliDurationSeconds` 耗时直方图
- `MeiliErrorsTotal` 错误计数
- `MeiliUpsertedDocsTotal` 写入文档数
- `MeiliDocumentsTotal` 文档总量（在 `DocumentCount` 时更新）

### 7. server.go — 独立 Metrics HTTP Server

```go
func NewMetricsServer(addr string, gatherer prometheus.Gatherer) *http.Server {
  mux := http.NewServeMux()
  mux.Handle("/metrics", promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}))
  return &http.Server{
    Addr:              addr,
    Handler:           mux,
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       10 * time.Second,
    WriteTimeout:      10 * time.Second,
  }
}
```

## 接口提取

### search.IndexOperator（新接口，在 meili_index.go 中定义）

```go
type IndexOperator interface {
  UpsertDocuments(ctx context.Context, docs []models.IndexDocument) error
  DeleteDocuments(ctx context.Context, docIDs []string) error
  DeleteAllDocuments(ctx context.Context) error
  DocumentCount(ctx context.Context) (int64, error)
  Search(params models.LocalSearchParams) ([]models.IndexDocument, int64, error)
  Ping() error
  EnsureSettings(ctx context.Context) error
}
```

`MeiliIndex` 已满足此接口，`InstrumentedMeiliIndex` 也实现此接口。

## echoprometheus 集成

在 `internal/httpx/server.go` 的 `NewServer` 中：

```go
func NewServer(handlers *Handlers, adminAPIKey string, distFS fs.FS, promReg prometheus.Registerer) *echo.Echo {
  e.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
    Subsystem:  "npan",
    Registerer: promReg,
    Skipper: func(c echo.Context) bool {
      p := c.Path()
      return p == "/healthz" || p == "/readyz"
    },
    LabelFuncs: map[string]echoprometheus.LabelValueFunc{
      "url": func(c echo.Context, err error) string {
        p := c.Path()
        if p == "/*" || p == "" {
          return "/spa"
        }
        return p
      },
    },
    HistogramOptsFunc: func(opts prometheus.HistogramOpts) prometheus.HistogramOpts {
      if opts.Name == "request_duration_seconds" {
        opts.Buckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0}
      }
      return opts
    },
  }))
  // ... 其余路由注册 ...
}
```

## Config 变更

```go
type Config struct {
  // ... 现有字段 ...
  MetricsAddr string // env: METRICS_ADDR, default: ":9091", 空字符串禁用
}
```

## main.go 组装顺序

1. 加载 Config
2. `metrics.NewRegistry()` → promReg
3. `metrics.NewSyncMetrics(promReg)` → syncMetrics
4. `metrics.NewSearchMetrics(promReg)` → searchMetrics
5. 创建 MeiliIndex → 包装为 InstrumentedMeiliIndex
6. 创建 QueryService → CachedQueryService → 包装为 InstrumentedSearchService
7. 创建 PrometheusSyncReporter → 注入 SyncManager
8. `httpx.NewServer(handlers, adminAPIKey, distFS, promReg)`
9. 启动主 HTTP server + metrics server
10. 优雅关闭：主 server 先关（15s），metrics server 后关（5s）

## Dockerfile 变更

```dockerfile
EXPOSE 1323 9091
```

## docker-compose.yml 变更

```yaml
services:
  npan:
    ports:
      - "1323:1323"
      - "9091:9091"
    environment:
      - METRICS_ADDR=:9091
```
