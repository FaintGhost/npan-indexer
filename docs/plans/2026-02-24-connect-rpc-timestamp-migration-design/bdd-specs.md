# BDD Specifications

## Feature: Connect 提供 Timestamp 语义且保持兼容

### Scenario: Connect progress 返回新 Timestamp 字段

```gherkin
Given 服务端已在 SyncProgressState 与相关子结构中新增 *_ts 字段
When 客户端调用 Connect 的 GetSyncProgress
Then 响应中应包含 Timestamp 字段
And Timestamp 与旧 int64 字段表示同一时刻
```

### Scenario: 前端优先读取 Timestamp 并正确展示

```gherkin
Given 前端收到仅包含 *_ts 的 progress 数据
When 页面渲染同步进度与时间信息
Then 应正确解析并展示时间
And 不应出现 NaN、空白或异常格式
```

### Scenario: 旧字段回退路径持续可用

```gherkin
Given 前端收到不包含 *_ts、仅包含旧 int64 字段的 progress 数据
When 页面渲染同步进度与时间信息
Then 应通过回退逻辑正确展示时间
And 行为与迁移前保持一致
```

## Feature: 存储与业务语义不回归

### Scenario: 进度持久化结构保持兼容

```gherkin
Given 进度存储仍以旧 int64 字段持久化
When 同步过程写入并读取进度
Then 读写应保持成功
And 不应要求旧存储文件迁移
```

### Scenario: 生成链路与回归验证通过

```gherkin
Given proto 新增 Timestamp sidecar 字段
When 执行 buf lint、buf generate 与测试回归
Then 生成产物应更新且构建通过
And 不应引入 REST/Connect 行为回归
```

## Suggested Automated Tests

## Go

- `internal/httpx`:
  - Connect progress 响应映射测试（新旧字段一致性）
  - descriptor/消息字段存在性测试（防止生成遗漏）
- `internal/service` / `internal/storage`:
  - 进度读写兼容回归（保持 int64 路径）

## Frontend

- `web/src/hooks/*`:
  - Timestamp 优先解析 + 旧字段回退
- `web/src/components/*`:
  - 时间展示在双输入场景下都正确

## Verification

- `buf lint`
- `buf generate`
- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./internal/httpx -count=1`
- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod go test ./... -count=1`
- `cd web && bun vitest run`（若涉及前端改动）
