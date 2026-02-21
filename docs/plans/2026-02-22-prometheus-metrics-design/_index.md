# Prometheus Metrics 设计

## 背景

npan 当前缺乏系统化的可观测性，仅有结构化日志和健康检查端点。为支持生产环境监控和告警，需要引入 Prometheus 指标体系。

## 需求

1. **HTTP 请求指标** — 请求总量、延迟直方图、错误率，按路由/方法/状态码分组
2. **同步任务指标** — 任务计数、耗时、文件处理数、失败数、当前状态
3. **Meilisearch/搜索指标** — 查询延迟、缓存命中/未命中、文档数量
4. **Go runtime 指标** — goroutine 数、内存分配、GC 暂停等
5. **独立端口** — metrics 端点在独立的内部端口上提供，不与主 API 共享

## 技术选型

| 组件 | 选择 | 理由 |
|------|------|------|
| HTTP 请求指标 | `echo-contrib/echoprometheus` | Echo v5 官方中间件，零配置提供 4 个标准 HTTP 指标 |
| 自定义业务指标 | `prometheus/client_golang` | Go 生态标准 Prometheus 客户端库 |
| Registry | `prometheus.NewRegistry()` | 自定义 registry，避免全局状态污染，利于测试 |
| Metrics 端口 | `:9091` | 避免与 Prometheus 自身 9090 冲突 |

## 设计原则

- **装饰器模式** — 与现有 `CachedQueryService` 装饰 `QueryService` 的模式保持一致，通过接口装饰器添加指标，不修改业务逻辑内部
- **依赖注入** — 指标系统在 `main.go` 中构建，通过构造函数注入各组件，与现有 DI 模式一致
- **零侵入可选** — `MetricsAddr` 为空时完全禁用 metrics，不影响其他功能
- **低基数标签** — 所有 label 唯一值不超过 100，避免高基数时序爆炸

## 指标一览

### HTTP 请求指标（echoprometheus 自动生成）

| 指标名 | 类型 | 标签 |
|--------|------|------|
| `npan_requests_total` | Counter | `code`, `method`, `host`, `url` |
| `npan_request_duration_seconds` | Histogram | `code`, `method`, `host`, `url` |
| `npan_response_size_bytes` | Histogram | `code`, `method`, `host`, `url` |
| `npan_request_size_bytes` | Histogram | `code`, `method`, `host`, `url` |

### 同步任务指标

| 指标名 | 类型 | 标签 |
|--------|------|------|
| `npan_sync_tasks_total` | Counter | `mode`, `status` |
| `npan_sync_duration_seconds` | Histogram | `mode` |
| `npan_sync_files_indexed_total` | Counter | `mode` |
| `npan_sync_files_failed_total` | Counter | `mode` |
| `npan_sync_running` | Gauge | — |
| `npan_sync_incremental_changes_total` | Counter | `op` |

### 搜索/Meilisearch 指标

| 指标名 | 类型 | 标签 |
|--------|------|------|
| `npan_search_queries_total` | Counter | `result` |
| `npan_search_cache_size` | Gauge | — |
| `npan_meili_operation_duration_seconds` | Histogram | `op` |
| `npan_meili_operation_errors_total` | Counter | `op` |
| `npan_meili_documents_total` | Gauge | — |
| `npan_meili_upserted_docs_total` | Counter | — |

### Go Runtime 指标（标准收集器）

- `go_goroutines`, `go_memstats_*`, `go_gc_duration_seconds`
- `process_cpu_seconds_total`, `process_resident_memory_bytes`, `process_open_fds`

## 文件变更清单

### 新增文件

| 文件 | 职责 |
|------|------|
| `internal/metrics/registry.go` | 创建自定义 Prometheus Registry，注册 Go/Process 收集器 |
| `internal/metrics/sync_metrics.go` | SyncMetrics 结构体及注册 |
| `internal/metrics/search_metrics.go` | SearchMetrics 结构体及注册 |
| `internal/metrics/sync_reporter.go` | SyncReporter 接口 + PrometheusSyncReporter 实现 |
| `internal/metrics/search_instrumenter.go` | InstrumentedSearchService 装饰器（缓存命中/未命中计数） |
| `internal/metrics/meili_instrumenter.go` | InstrumentedMeiliIndex 装饰器（操作耗时/错误计数） |
| `internal/metrics/server.go` | 独立 metrics HTTP server 启动/关闭 |

### 修改文件

| 文件 | 变更 |
|------|------|
| `internal/config/config.go` | 添加 `MetricsAddr` 字段 |
| `internal/config/validate.go` | `LogValue()` 中添加 MetricsAddr |
| `internal/httpx/server.go` | `NewServer` 增加 `promReg` 参数，注册 echoprometheus 中间件 |
| `internal/service/sync_manager.go` | `SyncManagerArgs` 添加 `MetricsReporter` 接口，在同步开始/结束时调用 |
| `internal/search/meili_index.go` | 提取 `IndexOperator` 接口 |
| `cmd/server/main.go` | 构建指标系统，启动独立 metrics server，并行优雅关闭 |
| `go.mod` | 添加 `prometheus/client_golang`、`echo-contrib` 依赖 |
| `Dockerfile` | 添加 `EXPOSE 9091` |
| `docker-compose.yml` | 添加 `9091:9091` 端口映射和 `METRICS_ADDR` 环境变量 |

## 数据流

```
HTTP 请求 → Echo (port 1323)
  │
  ├─ echoprometheus 中间件 → npan_requests_total / npan_request_duration_seconds
  │
  ├─ /api/v1/app/search → InstrumentedSearchService.Query()
  │    ├─ npan_search_queries_total{result=hit/miss}
  │    └─ CachedQueryService → QueryService → InstrumentedMeiliIndex.Search()
  │         └─ npan_meili_operation_duration_seconds{op=search}
  │
  └─ /api/v1/admin/sync → SyncManager.Start()
       ├─ npan_sync_running = 1
       ├─ ... 爬取/同步 ...
       │    └─ InstrumentedMeiliIndex.UpsertDocuments()
       │         └─ npan_meili_operation_duration_seconds{op=upsert}
       └─ npan_sync_running = 0
            └─ npan_sync_tasks_total{mode,status}

独立 Metrics Server (port 9091)
  └─ GET /metrics → promhttp.HandlerFor(registry)
```

## 优雅关闭顺序

1. 收到 SIGTERM/SIGINT
2. 先关闭主 HTTP 服务器（停止接收新业务请求，15s 超时）
3. 再关闭 Metrics 服务器（让 Prometheus 能抓取最终状态，5s 超时）

## Design Documents

- [BDD Specifications](./bdd-specs.md) - 行为场景和测试策略
- [Architecture](./architecture.md) - 系统架构和组件详情
- [Best Practices](./best-practices.md) - 安全、性能和代码质量指南
