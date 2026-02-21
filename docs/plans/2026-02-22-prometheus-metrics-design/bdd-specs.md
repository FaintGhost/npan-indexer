# BDD Specifications

## Feature: Metrics 端点可用性

```gherkin
Feature: Prometheus 指标服务器
  作为运维人员
  我需要一个独立的内部 metrics 端口
  以便 Prometheus 可以抓取指标而不暴露给公网

  Scenario: Metrics 端点在独立端口上可用
    Given 服务以主端口 ":1323" 和 Metrics 端口 ":9091" 启动
    When 客户端向 "http://localhost:9091/metrics" 发送 GET 请求
    Then 响应状态码应为 200
    And 响应 Content-Type 应包含 "text/plain"
    And 响应体应包含 "go_goroutines"

  Scenario: Metrics 端点不在主业务端口上暴露
    Given 服务以主端口 ":1323" 和 Metrics 端口 ":9091" 启动
    When 客户端向 "http://localhost:1323/metrics" 发送 GET 请求
    Then 响应状态码应为 404

  Scenario: MetricsAddr 为空时禁用 metrics
    Given 环境变量 METRICS_ADDR 为空
    When 服务启动
    Then 不应启动 metrics HTTP server
    And 主服务器正常运行
```

## Feature: HTTP 请求指标

```gherkin
Feature: HTTP 请求指标收集
  作为开发人员
  我需要 HTTP 请求的基础可观测性数据
  以便监控 API 延迟和错误率

  Scenario: 成功的 API 请求被计入请求总数
    Given Prometheus 中间件已启用
    When 客户端向 "/api/v1/search/local?query=test" 发送 GET 请求并返回 200
    Then "npan_requests_total" counter 应递增，标签含 method="GET", code="200"

  Scenario: 请求延迟被记录到 Histogram
    Given Prometheus 中间件已启用
    When 客户端发送请求
    Then "npan_request_duration_seconds" histogram 应记录观测值

  Scenario: 健康检查路由不被统计
    Given Prometheus 中间件已启用且配置了 Skipper
    When 客户端向 "/healthz" 发送 GET 请求
    Then "npan_requests_total" 中不应出现 url="/healthz" 的时序

  Scenario: 就绪检查路由不被统计
    Given Prometheus 中间件已启用且配置了 Skipper
    When 客户端向 "/readyz" 发送 GET 请求
    Then "npan_requests_total" 中不应出现 url="/readyz" 的时序

  Scenario: SPA 路由被规范化
    Given Prometheus 中间件配置了 LabelFunc
    When 客户端访问前端路由 "/settings"
    Then 对应 url label 应为 "/spa"
```

## Feature: 同步任务指标

```gherkin
Feature: 同步任务指标记录
  作为运维人员
  我需要同步任务的运行状态指标
  以便监控同步健康状况和设置告警

  Scenario: 同步启动时 running gauge 变为 1
    Given 当前无同步任务运行
    When 管理员发起同步任务
    Then "npan_sync_running" 应立即变为 1

  Scenario: 全量同步完成后指标正确更新
    When 全量同步成功完成，索引了 500 个文件
    Then "npan_sync_tasks_total{mode='full',status='done'}" 应递增 1
    And "npan_sync_files_indexed_total{mode='full'}" 应增加 500
    And "npan_sync_duration_seconds" 应记录实际耗时
    And "npan_sync_running" 应变回 0

  Scenario: 增量同步完成后指标正确更新
    When 增量同步成功完成
    Then "npan_sync_tasks_total{mode='incremental',status='done'}" 应递增 1
    And "npan_sync_running" 应变回 0

  Scenario: 同步被取消时状态为 cancelled
    Given 一个同步任务正在运行
    When 管理员取消同步
    Then "npan_sync_tasks_total{mode='...',status='cancelled'}" 应递增 1

  Scenario: 同步失败时状态为 error
    Given Meilisearch 不可用
    When 触发同步任务并失败
    Then "npan_sync_tasks_total{mode='...',status='error'}" 应递增 1
    And "npan_sync_files_failed_total" 应递增
```

## Feature: 搜索/缓存指标

```gherkin
Feature: 搜索与缓存指标记录
  作为开发人员
  我需要搜索层的缓存效率数据
  以便优化缓存配置

  Scenario: 首次搜索记录为缓存未命中
    Given 搜索缓存为空
    When 用户搜索关键词
    Then "npan_search_queries_total{result='miss'}" 应递增 1

  Scenario: 相同参数的二次搜索记录为缓存命中
    Given 搜索结果已在缓存中
    When 用户再次搜索相同关键词
    Then "npan_search_queries_total{result='hit'}" 应递增 1

  Scenario: Meilisearch 操作耗时被记录
    When 执行一次 Meilisearch search 操作
    Then "npan_meili_operation_duration_seconds{op='search'}" 应有新的观测值

  Scenario: Meilisearch 操作错误被计数
    Given Meilisearch 返回错误
    When 执行操作
    Then "npan_meili_operation_errors_total{op='...'}" 应递增

  Scenario: 文档总量 gauge 反映真实值
    Given Meilisearch 中有 N 个文档
    When 调用 DocumentCount
    Then "npan_meili_documents_total" 应设为 N
```

## Feature: Go Runtime 指标

```gherkin
Feature: Go 运行时指标
  作为运维人员
  我需要 Go runtime 指标
  以便监控内存和 goroutine 泄漏

  Scenario: 标准 Go 指标可在 metrics 端点获取
    When 访问 metrics 端点
    Then 响应体应包含 "go_goroutines"
    And 响应体应包含 "go_memstats_alloc_bytes"
    And 响应体应包含 "go_gc_duration_seconds"
    And 响应体应包含 "process_cpu_seconds_total"
```

## Feature: 优雅关闭

```gherkin
Feature: Metrics 服务器优雅关闭
  作为运维人员
  我需要 metrics 服务器在主服务器之后关闭
  以便 Prometheus 能抓取最终状态

  Scenario: 主服务器先于 metrics 服务器关闭
    Given 服务正在运行
    When 进程收到 SIGTERM
    Then 主服务器应先关闭
    And metrics 服务器应随后关闭

  Scenario: 关闭有超时保护
    Given 服务正在运行
    When 进程收到 SIGTERM
    Then 主服务器关闭超时为 15 秒
    And metrics 服务器关闭超时为 5 秒
```

## 测试策略

### 单元测试

使用 `prometheus/client_golang/prometheus/testutil` 包：

- **`testutil.ToFloat64(metric)`** — 读取单个指标当前值
- **`testutil.CollectAndCompare(collector, reader)`** — 精确文本比对
- **`testutil.GatherAndCompare(gatherer, reader)`** — 对整个 registry 比对

核心原则：每个测试使用独立的 `prometheus.NewRegistry()`，避免全局状态污染。

### 测试覆盖范围

1. `sync_metrics_test.go` — 验证 SyncMetrics 各计数器/直方图的递增行为
2. `search_metrics_test.go` — 验证缓存命中/未命中计数、Meilisearch 操作耗时记录
3. `sync_reporter_test.go` — 验证 PrometheusSyncReporter 正确转换 SyncEvent 到指标
4. `meili_instrumenter_test.go` — 验证装饰器正确记录耗时和错误
5. `server_test.go` — 验证 metrics HTTP server 端点可用性
6. `httpx/server_test.go` — 验证 echoprometheus 中间件的 Skipper 和 LabelFunc 行为
