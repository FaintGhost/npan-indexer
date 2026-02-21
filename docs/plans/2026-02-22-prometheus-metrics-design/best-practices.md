# Best Practices

## 命名规范

- Namespace: `npan`（全部指标统一前缀）
- Counter 必须以 `_total` 结尾
- 延迟指标使用 `_seconds` 后缀（基础单位）
- 字节指标使用 `_bytes` 后缀
- 全部 `snake_case` 小写
- 不将 label 名称写入 metric 名称

## Label 基数控制

**严禁使用的高基数 label：**
- `root_folder_id`、`query`（搜索词）、`user_id`、`request_id`
- 原始 URL 路径（需通过 LabelFunc 规范化）

**每个 label 唯一值上限：** 100 个

**当前设计的 label 基数分析：**

| Label | 值域 | 基数 |
|-------|------|------|
| `mode` | full, incremental | 2 |
| `status` | done, error, cancelled | 3 |
| `op` (sync) | upsert, delete, skip_upsert, skip_delete | 4 |
| `op` (meili) | search, upsert, delete, document_count | 4 |
| `result` | hit, miss | 2 |
| `method` | GET, POST, DELETE | 3 |
| `code` | 200, 400, 401, 404, 429, 500, 503 | ~7 |
| `url` | /api/v1/..., /spa | ~12 |

所有 label 均为低基数，安全。

## Histogram Bucket 设计

### HTTP 请求延迟

```
{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0}
```

覆盖：缓存命中(5ms) → 本地搜索(50ms) → 远程搜索(500ms) → 慢请求(10s)

### 同步任务耗时

```
{1, 5, 30, 60, 300, 600, 1800}
```

覆盖：小增量(1s) → 中等增量(30s) → 全量(30min)

### Meilisearch 操作耗时

```
{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5}
```

覆盖：快速查询(5ms) → 批量写入(2.5s)

## Registry 最佳实践

- 使用 `prometheus.NewRegistry()` 而非 `prometheus.DefaultRegisterer`
- 优势：测试隔离、避免第三方库冲突、精确控制暴露指标
- 通过 `promhttp.HandlerFor(gatherer, ...)` 暴露自定义 registry

## 安全考量

- Metrics 端点在独立内部端口（`:9091`），不与主 API 共享
- 不暴露敏感业务数据（不含文件名、用户名、搜索词等）
- 通过网络层（防火墙/安全组）限制 metrics 端口访问，仅允许 Prometheus 采集
- metrics 端点无认证（内部端口，依赖网络隔离）

## 性能考量

- Prometheus 指标操作（Inc、Observe 等）均为原子操作，开销极低（~100ns）
- Go runtime collector 和 process collector 开销约 2-5MB 基础内存
- 每条时序约 1KB 内存，当前设计约 50-100 条时序，总计 <1MB
- `GOMEMLIMIT=512MiB`（Dockerfile 中已设置）完全可容纳
- echoprometheus 中间件不影响请求处理延迟（ns 级别操作）

## 可选的 PromQL 告警规则参考

```yaml
# HTTP 5xx 错误率 > 5%（5分钟窗口）
- alert: HighErrorRate
  expr: rate(npan_requests_total{code=~"5.."}[5m]) / rate(npan_requests_total[5m]) > 0.05

# 同步任务连续失败
- alert: SyncFailures
  expr: increase(npan_sync_tasks_total{status="error"}[1h]) > 3

# 搜索缓存命中率 < 50%
- alert: LowCacheHitRate
  expr: rate(npan_search_queries_total{result="hit"}[5m]) / rate(npan_search_queries_total[5m]) < 0.5

# Goroutine 泄漏
- alert: GoroutineLeak
  expr: go_goroutines > 1000

# 同步任务运行超过 1 小时
- alert: SyncStuck
  expr: npan_sync_running == 1 and (time() - npan_sync_running) > 3600
```
