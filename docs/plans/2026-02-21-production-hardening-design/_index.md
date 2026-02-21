# npan 生产化加固设计

## 上下文

npan 是一个 Go 语言服务（Echo v5 + Meilisearch），作为 Novastar 内网云盘（Npan）的搜索代理。目前处于原型阶段，需要全面升级为可公网部署的生产级服务。

### 当前状态

- 单文件 HTML Demo 前端，挂载在 `/demo`
- 安全审计发现 13 项问题（3 Critical / 4 High / 6 Medium）
- 无认证中间件、无速率限制、无 CORS、无 body limit
- 配置无启动校验，AdminAPIKey 默认为空
- 无 Dockerfile、无健康检查增强

### 目标状态

- 部署于反向代理之后，TLS 由代理层处理
- 内嵌前端从 Demo 升级为生产级，路径从 `/demo` 迁移到 `/app`
- 双路径认证：内嵌前端自动使用服务端凭据，外部 API 消费者需提供 `X-API-Key`
- 全面消除安全审计发现
- 具备完整运维可观测性

## 用户决策

1. **部署模式**: 反向代理暴露到公网，Go 服务纯 HTTP
2. **前端定位**: 内嵌 HTML 保留并提升为生产级，路径 `/demo` → `/app`
3. **认证模型**: 内部/外部双路径
4. **改进范围**: 安全加固 + API 健壮性 + 运维能力 + API 重设计（全选）

## 需求列表

### Security（安全加固）

| # | 需求 | 优先级 | 说明 |
|---|------|--------|------|
| S-01 | 清除 Git 中的硬编码密钥 | P0 | 轮换凭据，确认 .env 未被 git 追踪 |
| S-02 | 实现双路径认证中间件 | P0 | 替代 per-handler 的 `requireAPIAccess()` |
| S-03 | 禁止空默认 AdminAPIKey | P0 | 启动校验，空 key 拒绝启动 |
| S-04 | 移除 query parameter token | P1 | 仅允许 Header/Body 传递凭据 |
| S-05 | pageSize 上限校验 | P1 | max 100，超出返回 400 |
| S-06 | 防御 Meilisearch filter 注入 | P1 | type 参数白名单 + 引号包裹 |
| S-07 | 防御 checkpoint 路径注入 | P1 | 路径规范化，限制在 data/checkpoints 下 |
| S-08 | 错误响应脱敏 | P2 | 内部日志记录完整错误，客户端返回通用消息 |
| S-09 | 进度接口脱敏 | P2 | 创建响应 DTO，排除 MeiliHost 等内部配置 |
| S-10 | 速率限制 | P2 | per-IP 限流，公开端点 20 req/s |
| S-11 | CORS 策略 | P2 | 仅允许自身 origin |
| S-12 | 请求体大小限制 | P2 | 全局 1MB |
| S-13 | 上游响应读取限制 | P2 | io.LimitReader 限制为 10MB |

### API Design（API 设计）

| # | 需求 | 优先级 | 说明 |
|---|------|--------|------|
| A-01 | 重新设计路由结构 | P0 | 公开/认证/管理三层路由组 |
| A-02 | 统一错误响应格式 | P1 | 含 code + message + request_id |
| A-03 | 输入参数校验 | P1 | type 白名单、数值范围、路径安全 |
| A-04 | 全局错误处理器 | P1 | 捕获 panic，返回标准化 500 |

### Ops（运维能力）

| # | 需求 | 优先级 | 说明 |
|---|------|--------|------|
| O-01 | 启动时配置验证 | P1 | 校验必填项、格式、安全性 |
| O-02 | 增强健康检查 | P1 | 新增 /readyz（含 Meili 连通性） |
| O-03 | 生产级 Dockerfile | P1 | 多阶段构建，非 root 运行 |
| O-04 | 优雅停机 | P1 | signal.NotifyContext + server.Shutdown |
| O-05 | 安全头中间件 | P2 | X-Content-Type-Options, X-Frame-Options, CSP |

### Frontend（前端）

| # | 需求 | 优先级 | 说明 |
|---|------|--------|------|
| F-01 | 路径迁移 /demo → /app | P0 | 旧路径 301 重定向 |
| F-02 | 前端生产化升级 | P1 | 移除 demo 标识，完善 UI |
| F-03 | 适配新认证模型 | P0 | 前端无需携带 API Key |

## 验收标准

- 无 Key 访问 `/app` 可正常使用搜索和下载
- 无 `X-API-Key` 头调用 `/api/v1/search/*` 返回 401
- API Key 为空时服务拒绝启动
- `pageSize=10000` 返回 400
- 注入 `type=file OR is_deleted = true` 返回 400
- 故意触发内部错误时，响应体不含堆栈或内部路径
- 高频请求触发 429
- `docker build` 成功，镜像 < 50MB，非 root 运行
- `GET /readyz` 在 Meili 不可达时返回 503
- `go test ./...` 和 `go test -race ./...` 全部通过

## 约束（不可变更）

1. 核心同步逻辑不变（`internal/indexer`）
2. Meilisearch 集成不变（`internal/search`）
3. CLI 接口不变（`cmd/cli`）
4. Npan API 客户端核心不变（`internal/npan`）
5. 存储层不变（`internal/storage`）
6. 技术栈不变（Go + Echo v5 + Meilisearch）
7. 单文件前端架构不变（内联 CSS/JS）

## 不在范围内

- 用户管理系统（仅静态 API Key）
- TLS 配置（由代理层处理）
- 数据库迁移（保持 JSON 文件存储）
- CI/CD 流水线
- 缓存层（Redis 等）
- 监控告警集成（Prometheus 等）
- 多租户支持

## 详细设计

详细设计参见以下文档。

## Design Documents

- [BDD Specifications](./bdd-specs.md) - 行为驱动规范（Gherkin 场景）
- [Architecture](./architecture.md) - 系统架构与组件设计
- [Best Practices](./best-practices.md) - 安全最佳实践与修复方案
