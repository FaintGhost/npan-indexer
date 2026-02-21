# 安全最佳实践与修复方案

## 1. HTTP 安全头

服务运行于反向代理之后，HSTS 由代理层处理。服务本身应设置：

```go
func secureHeaders() echo.MiddlewareFunc {
  return func(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c *echo.Context) error {
      h := c.Response().Header()
      h.Set("X-Content-Type-Options", "nosniff")
      h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
      h.Set("X-Frame-Options", "DENY")
      h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
      return next(c)
    }
  }
}
```

对前端页面端点额外设置 CSP：

```go
h.Set("Content-Security-Policy",
  "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; "+
  "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; "+
  "font-src 'self' https://fonts.gstatic.com; "+
  "connect-src 'self'; frame-ancestors 'none'")
```

## 2. 凭据管理策略

### 分层方案

| 环境 | 方案 |
|------|------|
| 本地开发 | `.env.example` + 本地 `.env`（gitignored） |
| CI/CD | GitHub Actions Secrets / 环境变量 |
| 生产 | Docker Secrets 或环境变量注入，不落盘 |

### Git 历史清理

```bash
# 确认文件是否被追踪
git ls-files --error-unmatch .env .env.meilisearch 2>&1

# 如果被追踪，移除并提交
git rm --cached .env .env.meilisearch
git commit -m "chore: remove secret files from tracking"

# 轮换所有已泄露凭据
# - OAuth client_id/client_secret
# - Meilisearch master key / API key
# - Admin API key
```

### Config 日志脱敏

```go
func (c Config) LogValue() slog.Value {
  return slog.GroupValue(
    slog.String("server_addr", c.ServerAddr),
    slog.String("base_url", c.BaseURL),
    slog.String("meili_host", c.MeiliHost),
    slog.String("admin_api_key", "[REDACTED]"),
    slog.String("client_secret", "[REDACTED]"),
    slog.String("meili_api_key", "[REDACTED]"),
    slog.String("token", "[REDACTED]"),
  )
}
```

## 3. 逐项修复方案

### #1 [Critical] 清除 Git 中的凭据

- 轮换所有泄露的凭据
- `git rm --cached .env .env.meilisearch`
- 确认 `.gitignore` 规则生效
- 考虑 `git filter-repo` 清理历史

### #2 [Critical] Demo 端点无认证

**路由重设计方案**：移除 `/api/v1/demo/*`，替换为 `/api/v1/app/*`。`/api/v1/app/*` 使用 `EmbeddedAuth` 中间件，自动注入 `allow_config_fallback=true`，无需外部 API Key。

**下载端点加固**：即使内嵌前端无需 API Key，仍应验证 file_id 在索引中存在，并限制 valid_period 上限。

### #3 [Critical] AdminAPIKey 默认为空

在 `cmd/server/main.go` 中 `config.Load()` 后调用 `cfg.Validate()`，空 Key 拒绝启动。

### #4 [High] 移除 Query Parameter Token

删除 `handlers.go:104` 中的 `c.QueryParam("token")`，仅保留 Header 方式传递。

### #5 [High] pageSize 上限

```go
const maxPageSize int64 = 100

// handler 层
if parsed > maxPageSize {
  return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest,
    fmt.Sprintf("page_size 不能超过 %d", maxPageSize))
}

// query_service.go 兜底
if normalized.PageSize > 100 {
  normalized.PageSize = 100
}
```

### #6 [High] Meilisearch Filter 注入

**双层防御**：

第一层 — handler 白名单：
```go
var allowedTypes = map[string]bool{
  "all": true, "file": true, "folder": true,
}
if !allowedTypes[typeParam] {
  return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest,
    "type 参数无效，允许值: all, file, folder")
}
```

第二层 — meili_index.go 引号包裹：
```go
filters = append(filters, fmt.Sprintf("type = '%s'", params.Type))
```

### #7 [High] Checkpoint 路径注入

```go
func validateCheckpointTemplate(template string) error {
  if template == "" {
    return nil
  }
  cleaned := filepath.Clean(template)
  if filepath.IsAbs(cleaned) {
    return fmt.Errorf("检查点路径无效: 不允许绝对路径")
  }
  if strings.Contains(cleaned, "..") {
    return fmt.Errorf("检查点路径无效: 不允许路径遍历")
  }
  if !strings.HasPrefix(cleaned, "data"+string(filepath.Separator)+"checkpoints") &&
     !strings.HasPrefix(cleaned, "data/checkpoints") {
    return fmt.Errorf("检查点路径无效: 必须在 data/checkpoints 目录下")
  }
  return nil
}
```

### #8 [Medium] 错误信息脱敏

所有 `writeError(c, status, err.Error())` 替换为：

```go
slog.Error("操作失败", "error", err, "handler", "Token")
return writeErrorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "操作失败，请检查参数")
```

具体映射表：

| 位置 | 当前 | 替换为 |
|------|------|--------|
| Token:208 | `err.Error()` | "认证失败，请检查凭据" |
| RemoteSearch:229 | `err.Error()` | "搜索请求失败，请稍后重试" |
| DownloadURL:284 | `err.Error()` | "获取下载链接失败" |
| LocalSearch:357 | `err.Error()` | "搜索服务暂不可用" |
| StartFullSync:474 | `err.Error()` | "启动同步失败" |
| GetProgress:490 | `err.Error()` | "无法读取同步进度" |

### #9 [Medium] 进度接口脱敏

创建 `SyncProgressResponse` DTO（见 architecture.md），在 `GetFullSyncProgress` handler 中做转换后再返回。

### #10 [Medium] 速率限制

基于 `golang.org/x/time/rate`（已在 go.mod 中）实现 per-IP 令牌桶：

| 端点类别 | 限流阈值 |
|----------|----------|
| 公开搜索 /app/* | 20 req/s, burst 40 |
| Token 端点 | 10 req/min |
| 管理搜索 | 60 req/s |
| Sync 操作 | 5 req/min |

### #11 [Medium] CORS

```go
e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
  AllowOrigins: []string{"https://your-domain.com"},
  AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
  AllowHeaders: []string{"Authorization", "X-API-Key", "Content-Type"},
  MaxAge:       3600,
}))
```

通过环境变量 `CORS_ALLOWED_ORIGINS` 配置。

### #12 [Medium] Body 大小限制

```go
e.Use(middleware.BodyLimit("1MB"))
```

### #13 [Medium] io.ReadAll 限制

```go
// client.go readStatusError
limited := io.LimitReader(resp.Body, 4096)
body, _ := io.ReadAll(limited)

// client.go request 正常响应
limited := io.LimitReader(resp.Body, 10*1024*1024) // 10MB
return json.NewDecoder(limited).Decode(out)
```

## 4. 其他改进

### 代码去重

`firstNotEmpty()` 和 `firstPositive()` 在 `handlers.go` 和 `cli/root.go` 中重复。`toInt64()` 和 `toBool()` 在 `npan/client.go` 和 `indexer/incremental_fetch.go` 中重复。建议抽取到 `internal/util` 包。

### 优雅停机

当前 `cmd/server/main.go` 直接调用 `server.Start()` 不处理信号，同步任务运行中收到 SIGTERM 会被强制终止。改为 `signal.NotifyContext` + `server.Shutdown(ctx)` 实现优雅关闭。

### Readyz 健康检查

```go
func (h *Handlers) Readyz(c *echo.Context) error {
  if err := h.queryService.Ping(); err != nil {
    return c.JSON(503, map[string]any{"status": "not_ready", "meili": "unreachable"})
  }
  return c.JSON(200, map[string]any{"status": "ready"})
}
```

### Handler 返回 nil 问题

当前 `requireAPIAccess` 返回 false 时 handler 返回 `nil`。改为中间件模式后此问题自动消除。
