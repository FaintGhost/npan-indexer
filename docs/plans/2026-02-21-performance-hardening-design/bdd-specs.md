# BDD Specifications — 2C2G 性能加固

## Feature 1: HTTP Server 超时配置

### Scenario 1.1: 服务启动时配置超时参数

```gherkin
Feature: HTTP Server 超时配置

  Scenario: 服务启动时 http.Server 配置了超时参数
    Given 应用通过 cmd/server/main.go 启动
    When http.Server 实例被创建
    Then ReadHeaderTimeout 应为 5s
    And ReadTimeout 应为 10s
    And WriteTimeout 应为 30s
    And IdleTimeout 应为 120s
```

### Scenario 1.2: 超时参数可通过配置覆盖

```gherkin
  Scenario: 超时参数可通过环境变量覆盖
    Given 环境变量 SERVER_READ_TIMEOUT=15s
    And 环境变量 SERVER_WRITE_TIMEOUT=60s
    When 应用启动并加载配置
    Then ReadTimeout 应为 15s
    And WriteTimeout 应为 60s
```

## Feature 2: 速率限制中间件挂载

### Scenario 2.1: 搜索端点速率限制

```gherkin
Feature: 速率限制

  Scenario: 搜索端点在超过限制时返回 429
    Given 速率限制为 20 req/s per IP, burst 40
    And 客户端 IP 为 "192.168.1.1"
    When 客户端在 1 秒内发送 50 个搜索请求
    Then 前 40 个请求返回 200
    And 后续请求返回 429
    And 响应头包含 "Retry-After"
```

### Scenario 2.2: 不同 IP 独立限流

```gherkin
  Scenario: 不同 IP 的限流器互不影响
    Given 速率限制为 20 req/s per IP
    And 客户端 A IP 为 "192.168.1.1"
    And 客户端 B IP 为 "192.168.1.2"
    When 客户端 A 耗尽速率限制
    Then 客户端 B 的请求仍然返回 200
```

### Scenario 2.3: 管理端点独立限流

```gherkin
  Scenario: 管理端点使用更严格的速率限制
    Given 管理端点速率限制为 5 req/s per IP, burst 10
    When 客户端在 1 秒内发送 15 个管理请求
    Then 前 10 个请求返回 200
    And 后续请求返回 429
```

## Feature 3: 搜索结果 LRU 缓存

### Scenario 3.1: 缓存命中

```gherkin
Feature: 搜索结果缓存

  Scenario: 相同搜索参数命中缓存
    Given CachedQueryService 已初始化，TTL=30s, capacity=256
    And 用户搜索 query="mx40 spec" type="file" page=1
    And Meilisearch 返回 5 个结果
    When 用户再次搜索相同参数
    Then 返回结果与第一次完全相同
    And Meilisearch 未被再次调用（缓存命中）
```

### Scenario 3.2: 缓存过期

```gherkin
  Scenario: 缓存条目在 TTL 后过期
    Given CachedQueryService 已初始化，TTL=30s
    And 用户搜索 query="mx40 spec"
    When 等待 31 秒后再次搜索相同参数
    Then Meilisearch 被再次调用（缓存未命中）
```

### Scenario 3.3: 不同参数独立缓存

```gherkin
  Scenario: 不同搜索参数使用不同缓存条目
    Given CachedQueryService 已初始化
    When 用户搜索 query="mx40 spec" type="file"
    And 用户搜索 query="mx40 spec" type="folder"
    Then Meilisearch 被调用 2 次（参数不同）
```

### Scenario 3.4: LRU 容量淘汰

```gherkin
  Scenario: 缓存满时淘汰最久未访问的条目
    Given CachedQueryService 已初始化，capacity=2
    And 缓存中有 query="a" 和 query="b"
    When 用户搜索 query="c"（新条目）
    Then query="a" 被淘汰（最久未访问）
    And query="b" 和 query="c" 仍在缓存中
```

## Feature 4: 同步动态降速

### Scenario 4.1: 搜索活跃时降低同步速率

```gherkin
Feature: 同步动态降速

  Scenario: 搜索活跃时同步降速
    Given SearchActivityTracker 跟踪搜索活动
    And 同步正在运行，正常速率为 N req/s
    When 最近 5 秒内有搜索请求
    Then 同步速率降低至 N/2 req/s
```

### Scenario 4.2: 搜索空闲时恢复同步速率

```gherkin
  Scenario: 搜索空闲后同步恢复正常速率
    Given 搜索活跃导致同步降速中
    When 连续 10 秒无搜索请求
    Then 同步速率恢复为正常值 N req/s
```

### Scenario 4.3: 搜索活动追踪器记录活动

```gherkin
  Scenario: 搜索请求被正确追踪
    Given SearchActivityTracker 初始化，窗口=5s
    When 一个搜索请求到来
    Then IsActive() 返回 true
    When 等待 6 秒后
    Then IsActive() 返回 false
```

## Feature 5: 运行时调优

### Scenario 5.1: Dockerfile 设置 GOMEMLIMIT

```gherkin
Feature: 运行时调优

  Scenario: Dockerfile 包含 GOMEMLIMIT 环境变量
    Given Dockerfile 生产镜像阶段
    When 检查环境变量配置
    Then GOMEMLIMIT 应设置为 "512MiB"
```

### Scenario 5.2: GOMEMLIMIT 可通过环境变量覆盖

```gherkin
  Scenario: 运行时可通过环境变量覆盖 GOMEMLIMIT
    Given Dockerfile 中 GOMEMLIMIT=512MiB
    When docker run -e GOMEMLIMIT=768MiB
    Then 实际 GOMEMLIMIT 为 768MiB
```

## 测试策略

### 单元测试

| 组件 | 测试方式 | 外部依赖隔离 |
|------|----------|-------------|
| CachedQueryService | 注入 mock QueryService | mock 替代 Meilisearch |
| SearchActivityTracker | 时间控制测试 | 无外部依赖 |
| HTTP 超时配置 | 检查 http.Server 字段值 | 无外部依赖 |
| 速率限制挂载 | Echo test request | 无外部依赖 |

### 集成验证

- 启动服务后检查 `http.Server` 超时字段
- 高频请求触发 429 响应
- 搜索缓存命中减少 Meilisearch 调用
- Dockerfile 构建验证 `GOMEMLIMIT` 环境变量
