# 最佳实践 — 2C2G 性能加固

## HTTP Server 超时

### Go 标准库超时模型

Go 的 `http.Server` 提供四层超时保护:

```
连接建立 → ReadHeaderTimeout → ReadTimeout → Handler 执行 → WriteTimeout
                                                           ↓
                                                    IdleTimeout (keep-alive)
```

### 推荐值（2C2G 环境）

| 参数 | 值 | 依据 |
|------|----|------|
| ReadHeaderTimeout | 5s | 防止 slowloris 攻击，内网环境 5s 足够 |
| ReadTimeout | 10s | 搜索请求体很小，10s 覆盖慢网络 |
| WriteTimeout | 30s | Meilisearch 搜索 P99 < 500ms，30s 留大量余量 |
| IdleTimeout | 120s | 复用 keep-alive，减少连接开销 |

### 注意事项

- `WriteTimeout` 从 request header 读取完成开始计时，包含 handler 执行时间
- 如果 handler 需要长时间处理（如下载 URL 生成），`WriteTimeout` 必须覆盖
- 超时参数应可通过环境变量覆盖，方便不同部署环境调整

## 速率限制

### Token Bucket 算法

当前实现使用 `golang.org/x/time/rate`（token bucket），适合此场景:

- **rps**: 每秒填充的 token 数
- **burst**: 桶容量，允许瞬时突发
- 建议 burst = 2 * rps，允许合理突发

### 分层限流策略

```
全局层: 20 rps, burst 40  ← 覆盖所有端点
  └── 管理端点: 5 rps, burst 10  ← 更严格限制管理操作
```

### IP 清理

当前实现每分钟清理 3 分钟未见的 IP，适合内网场景。大规模部署需考虑:

- 使用 `sync.Map` 替代 `map + Mutex`（高并发下更优）
- 或使用 `golang-lru` 自带的 TTL 淘汰

## 搜索缓存

### 库选择: hashicorp/golang-lru/v2/expirable

| 库 | LRU | TTL | 线程安全 | 泛型 | 维护状态 |
|----|-----|-----|---------|------|---------|
| hashicorp/golang-lru/v2 | Yes | Yes (expirable) | Yes | Yes | 活跃 |
| patrickmn/go-cache | No | Yes | Yes | No | 停止维护 |
| bigcache | No | Yes | Yes | No | 适合大量小值 |

选择 `hashicorp/golang-lru/v2/expirable`:

- LRU + TTL 双重淘汰策略
- 泛型支持，类型安全
- 无需额外 goroutine 清理

### 缓存键设计

```go
func cacheKey(p models.LocalSearchParams) string {
    var b strings.Builder
    b.WriteString(p.Query)
    b.WriteByte('|')
    b.WriteString(p.Type)
    b.WriteByte('|')
    fmt.Fprintf(&b, "%d|%d", p.Page, p.PageSize)
    if p.ParentID != nil {
        fmt.Fprintf(&b, "|p%d", *p.ParentID)
    }
    if p.UpdatedAfter != nil {
        fmt.Fprintf(&b, "|a%d", *p.UpdatedAfter)
    }
    if p.UpdatedBefore != nil {
        fmt.Fprintf(&b, "|b%d", *p.UpdatedBefore)
    }
    if p.IncludeDeleted {
        b.WriteString("|d")
    }
    return b.String()
}
```

要点:
- 使用 `strings.Builder` 减少分配
- 字段之间用 `|` 分隔，避免歧义
- 指针字段带前缀标记，区分 nil 和零值

### 缓存一致性

- TTL 30 秒保证数据时效性（同步批次间隔通常 > 30s）
- 不主动失效缓存（同步写入后不清缓存），依赖 TTL 过期
- 理由: 搜索结果的短暂不一致可接受（文件索引场景，非交易系统）

## 同步动态降速

### 设计原则

- **非侵入**: 不修改 SyncManager 核心逻辑，仅调整 `rate.Limiter` 参数
- **自动恢复**: 搜索空闲后自动恢复，无需人工干预
- **平滑过渡**: 使用 `SetLimit()` 而非创建新 limiter，避免突变

### 活动窗口

- 窗口设为 5 秒: 搜索请求间隔通常 < 5 秒（用户连续搜索）
- 使用 `atomic.Int64` 存储时间戳，无锁高性能
- `IsActive()` 检查: `time.Now().Unix() - lastActive < windowSec`

### 降速策略

```
搜索活跃:
  同步速率 = 正常速率 × 0.5

搜索空闲:
  同步速率 = 正常速率 × 1.0
```

简单二级策略，避免复杂度。如未来需要更精细控制，可扩展为多级。

## Go 运行时调优

### GOMEMLIMIT

- **2C2G 共享部署**: Meilisearch ~1GB，系统 ~256MB，npan-server ~512MB
- 设置 `GOMEMLIMIT=512MiB` 让 GC 在接近上限时更积极回收
- 防止 Go 进程无限增长触发 OOM killer

### GOGC

- 默认值 `GOGC=100`（堆大小翻倍时触发 GC）
- 在有 `GOMEMLIMIT` 的情况下，`GOGC=100` 是合理默认
- 不建议降低 GOGC（如 50），因为会增加 CPU 压力（2C 已经紧张）

### Meilisearch 资源限制

建议在 docker-compose 或启动参数中配置:

```yaml
meilisearch:
  command:
    - --max-indexing-threads
    - "1"
    - --max-indexing-memory
    - "800Mb"
  deploy:
    resources:
      limits:
        memory: 1280M
```

- `--max-indexing-threads 1`: 限制索引只用 1 个 CPU 核心，留 1 核给搜索
- `--max-indexing-memory 800Mb`: 限制索引内存使用
- Docker memory limit 1280M: 留余量给搜索和运行时开销

## 性能预期

### 2C2G 同机部署参考数据

| 场景 | 预期 QPS | P95 延迟 |
|------|---------|---------|
| 纯搜索（缓存命中） | 500+ | < 1ms |
| 纯搜索（缓存未命中） | 50-100 | 10-50ms |
| 搜索 + 同步并行 | 30-80 | 20-100ms |
| 纯同步 | N/A | N/A |

### 内存预算

| 组件 | 预估内存 |
|------|---------|
| Meilisearch 运行时 | ~800 MB |
| Meilisearch 索引 | ~200 MB |
| npan-server 基础 | ~50 MB |
| LRU 缓存（256 条目） | ~2-4 MB |
| Go 运行时 + goroutine 栈 | ~50 MB |
| 系统 + 余量 | ~200 MB |
| **总计** | ~1.3-1.5 GB |

在 2 GB 总内存下，留有 500-700 MB 余量用于系统缓存和突发。
