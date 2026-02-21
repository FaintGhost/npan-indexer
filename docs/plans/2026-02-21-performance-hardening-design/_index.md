# npan 2C2G 性能加固设计

## 上下文

npan 是一个 Go 语言服务（Echo v5 + Meilisearch），作为 Novastar 内网云盘的搜索代理。生产化安全加固已完成，现需针对 **2C2G（2 vCPU / 2 GB RAM）同机部署**场景进行性能加固。

### 部署架构

- 单台 2C2G 服务器运行 npan-server + Meilisearch
- Meilisearch 预计消耗 ~1 GB RAM（索引数据）
- npan-server 可用内存约 512–768 MB
- 两个进程共享 2 个 CPU 核心

### 当前问题

| # | 问题 | 影响 |
|---|------|------|
| P-01 | `http.Server` 无超时配置 | 慢客户端占满连接池 |
| P-02 | `RateLimitMiddleware` 已实现但未挂载 | 恶意/异常流量无限制 |
| P-03 | 搜索无缓存 | 相同查询重复请求 Meilisearch |
| P-04 | 同步期间无动态降速 | 搜索+同步并行时 CPU/IO 争抢 |
| P-05 | 无 `GOMEMLIMIT` 配置 | GC 压力大，可能 OOM |

### 核心场景

**搜索 + 同步并行**：用户使用搜索功能时，后台可能正在执行全量/增量同步。2C2G 下两者争抢 CPU 和 Meilisearch 资源。

## 用户决策

1. **关注场景**: 搜索 + 同步并行
2. **部署模式**: 同机部署（npan-server + Meilisearch 同一台 2C2G 机器）
3. **加固方案**: 方案 A：轻量加固（HTTP 超时 + 限流 + 缓存 + 同步降速）

## 需求列表

### P0：HTTP Server 超时

| 参数 | 值 | 说明 |
|------|----|------|
| ReadHeaderTimeout | 5s | 防止 slowloris 攻击 |
| ReadTimeout | 10s | 限制请求体读取时间 |
| WriteTimeout | 30s | 覆盖搜索+下载 URL 生成 |
| IdleTimeout | 120s | 复用 keep-alive 连接 |

### P0：挂载速率限制中间件

- 在 `NewServer()` 中挂载已有的 `RateLimitMiddleware`
- 搜索端点: 20 req/s per IP, burst 40
- 管理端点: 5 req/s per IP, burst 10

### P1：搜索结果 LRU 缓存

- 使用 `hashicorp/golang-lru/v2/expirable` 实现带 TTL 的 LRU 缓存
- 缓存键: 序列化的 `LocalSearchParams`
- TTL: 30 秒（同步间隔内有效）
- 容量: 256 条目（约 2–4 MB 内存）
- 实现为 `CachedQueryService` 装饰器，包装 `QueryService`

### P2：同步动态降速

- 搜索活跃时（最近 N 秒内有搜索请求），降低同步并发
- 通过 `rate.Limiter.SetLimit()` 动态调整同步速率
- 搜索空闲时恢复正常同步速度

### P1：运行时调优

- Dockerfile 中设置 `GOMEMLIMIT=512MiB`
- Dockerfile 中设置 `GOGC=100`（显式声明默认值）
- Meilisearch docker-compose 建议: `--max-indexing-threads 1`, `--max-indexing-memory 800Mb`

## 验收标准

- `http.Server` 配置了 4 个超时参数
- 搜索端点在高频请求时返回 429
- 相同搜索参数在 30 秒内命中缓存（cache hit）
- 同步运行期间，搜索响应 P95 < 200ms
- `GOMEMLIMIT` 在 Dockerfile 中正确设置
- `go test ./...` 全部通过

## 约束

1. 不引入外部缓存服务（Redis 等）
2. 不改变同步核心逻辑（`internal/indexer`）
3. 不改变 Meilisearch 集成接口（`internal/search`）
4. 缓存必须进程内存实现
5. 同步降速必须自动/透明，无需人工干预

## 不在范围内

- Prometheus/Grafana 监控集成
- 分布式限流
- CDN / 反向代理缓存
- 连接池优化（当前 HTTP 客户端已有 30s timeout）
- 数据库层优化

## Design Documents

- [BDD Specifications](./bdd-specs.md) - 行为驱动规范（Gherkin 场景）
- [Architecture](./architecture.md) - 系统架构与组件设计
- [Best Practices](./best-practices.md) - 2C2G 环境最佳实践
