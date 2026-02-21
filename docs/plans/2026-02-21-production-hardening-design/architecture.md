# 系统架构

## 1. 当前架构

### 包结构与依赖流

```
cmd/server ─┬─> config
             ├─> search (MeiliIndex, QueryService)
             ├─> storage (JSONProgressStore)
             ├─> service (SyncManager)
             ├─> httpx (Handlers, NewServer)
             └─> logx

httpx ──┬─> config
        ├─> npan (API client, auth)
        ├─> search (QueryService)
        ├─> service (SyncManager, DownloadURLService)
        └─> models

service ──┬─> indexer (RunFullCrawl)
          ├─> search (MeiliIndex)
          ├─> storage
          └─> npan (API)

indexer ──┬─> npan (API)
          ├─> search (mapper)
          └─> models
```

### 当前中间件栈（仅 3 个）

```go
e.Use(middleware.RequestID())
e.Use(middleware.Recover())
e.Use(middleware.RequestLogger())
```

## 2. 目标架构

### 新路由结构

```
# 无认证
GET  /healthz                          存活检查（LB/K8s 探针）
GET  /readyz                           就绪检查（含 Meili 连通性）
GET  /app                              内嵌生产前端
GET  /app/*                            前端静态资源（预留）

# 公共端点（内嵌前端，服务端凭据，无需 API Key）
GET  /api/v1/app/search                本地搜索
GET  /api/v1/app/download-url          下载链接

# 外部 API（X-API-Key 必须）
POST /api/v1/token                     获取 Npan token
GET  /api/v1/search/remote             远程搜索
GET  /api/v1/search/local              本地搜索
GET  /api/v1/download-url              文件下载链接

# 管理端点（X-API-Key 必须）
POST /api/v1/admin/sync/full           启动全量同步
GET  /api/v1/admin/sync/full/progress  查看进度
POST /api/v1/admin/sync/full/cancel    取消同步
```

### 新中间件栈

```go
e.Use(middleware.RequestID())
e.Use(middleware.Recover())
e.Use(secureHeaders())                    // 新增: 安全头
e.Use(middleware.BodyLimit("1MB"))         // 新增: body 大小限制
e.Use(middleware.RequestLogger())
e.Use(rateLimitMiddleware(20, 40))        // 新增: 速率限制

// 路由组级别中间件
appGroup.Use(embeddedAuth())              // 新增: 内嵌前端自动凭据
apiGroup.Use(apiKeyAuth(cfg.AdminAPIKey)) // 新增: API Key 认证
adminGroup.Use(apiKeyAuth(cfg.AdminAPIKey))
```

## 3. 认证中间件设计

### 双路径模型

```
内嵌前端请求流:
  Browser → GET /app (HTML 页面)
  Browser → GET /api/v1/app/search?q=xxx
         → embeddedAuth 中间件自动注入 auth_mode=embedded
         → handler 检测 embedded 模式，使用服务端凭据

外部 API 调用:
  Client → GET /api/v1/search/local?q=xxx + X-API-Key header
         → apiKeyAuth 中间件验证 Key
         → handler 使用请求者提供的凭据
```

### 实现要点

```go
// internal/httpx/middleware_auth.go

func APIKeyAuth(adminKey string) echo.MiddlewareFunc {
  return func(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c *echo.Context) error {
      provided := strings.TrimSpace(c.Request().Header.Get("X-API-Key"))
      if provided == "" {
        provided = parseBearerHeader(c.Request().Header.Get("Authorization"))
      }
      if subtle.ConstantTimeCompare([]byte(provided), []byte(adminKey)) != 1 {
        return c.JSON(http.StatusUnauthorized, ErrorResponse{
          Code:    "UNAUTHORIZED",
          Message: "未授权：缺少或无效的 API Key",
        })
      }
      return next(c)
    }
  }
}

func EmbeddedAuth() echo.MiddlewareFunc {
  return func(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c *echo.Context) error {
      c.Set("auth_mode", "embedded")
      c.Set("allow_config_fallback", true)
      return next(c)
    }
  }
}
```

## 4. 统一错误处理

### ErrorResponse 结构

```go
// internal/httpx/errors.go

type ErrorResponse struct {
  Code      string `json:"code"`
  Message   string `json:"message"`
  RequestID string `json:"request_id,omitempty"`
}

// 预定义错误码
const (
  ErrCodeUnauthorized  = "UNAUTHORIZED"
  ErrCodeBadRequest    = "BAD_REQUEST"
  ErrCodeNotFound      = "NOT_FOUND"
  ErrCodeConflict      = "CONFLICT"
  ErrCodeRateLimited   = "RATE_LIMITED"
  ErrCodeInternalError = "INTERNAL_ERROR"
)

func writeErrorResponse(c *echo.Context, status int, code string, message string) error {
  return c.JSON(status, ErrorResponse{
    Code:    code,
    Message: message,
    RequestID: extractRequestID(c),
  })
}
```

### 全局错误处理器

```go
func customHTTPErrorHandler(c *echo.Context, err error) {
  var echoErr *echo.HTTPError
  if errors.As(err, &echoErr) {
    writeErrorResponse(c, echoErr.Code, "HTTP_ERROR", fmt.Sprintf("%v", echoErr.Message))
    return
  }
  slog.Error("未处理的服务器错误", "error", err, "request_id", extractRequestID(c))
  writeErrorResponse(c, 500, ErrCodeInternalError, "服务器内部错误")
}
```

## 5. 响应 DTO

### SyncProgressResponse（脱敏版）

```go
// internal/httpx/dto.go

type SyncProgressResponse struct {
  Status         string                           `json:"status"`
  StartedAt      int64                            `json:"startedAt"`
  UpdatedAt      int64                            `json:"updatedAt"`
  Roots          []int64                          `json:"roots"`
  CompletedRoots []int64                          `json:"completedRoots"`
  ActiveRoot     *int64                           `json:"activeRoot,omitempty"`
  AggregateStats models.CrawlStats                `json:"aggregateStats"`
  RootProgress   map[string]*RootProgressResponse `json:"rootProgress"`
  LastError      string                           `json:"lastError,omitempty"`
  // 刻意排除: MeiliHost, MeiliIndex, CheckpointTemplate
}

type RootProgressResponse struct {
  RootFolderID       int64             `json:"rootFolderId"`
  Status             string            `json:"status"`
  EstimatedTotalDocs *int64            `json:"estimatedTotalDocs,omitempty"`
  Stats              models.CrawlStats `json:"stats"`
  UpdatedAt          int64             `json:"updatedAt"`
  // 刻意排除: CheckpointFile, Error 具体内容
}
```

## 6. 配置验证

```go
// internal/config/validate.go

func (c Config) Validate() error {
  var errs []string

  // 安全必填项
  if strings.TrimSpace(c.AdminAPIKey) == "" {
    errs = append(errs, "NPA_ADMIN_API_KEY 不能为空")
  } else if len(c.AdminAPIKey) < 16 {
    errs = append(errs, "NPA_ADMIN_API_KEY 长度不应少于 16 字符")
  }

  // 基础必填项
  if c.MeiliHost == "" {
    errs = append(errs, "MEILI_HOST 不能为空")
  }
  if c.MeiliIndex == "" {
    errs = append(errs, "MEILI_INDEX 不能为空")
  }
  if c.BaseURL == "" {
    errs = append(errs, "NPA_BASE_URL 不能为空")
  }

  // 数值范围
  if c.SyncMaxConcurrent <= 0 || c.SyncMaxConcurrent > 20 {
    errs = append(errs, "NPA_SYNC_MAX_CONCURRENT 应在 1-20 之间")
  }
  if c.Retry.MaxRetries < 0 || c.Retry.MaxRetries > 10 {
    errs = append(errs, "NPA_MAX_RETRIES 应在 0-10 之间")
  }

  // 认证完整性
  hasClientCreds := c.ClientID != "" && c.ClientSecret != "" && c.SubID > 0
  hasToken := c.Token != ""
  if c.AllowConfigAuthFallback && !hasClientCreds && !hasToken {
    errs = append(errs, "NPA_ALLOW_CONFIG_AUTH_FALLBACK=true 但未提供有效凭据")
  }

  if len(errs) > 0 {
    return fmt.Errorf("配置验证失败:\n  - %s", strings.Join(errs, "\n  - "))
  }
  return nil
}
```

## 7. Dockerfile

```dockerfile
# Stage 1: Build
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-s -w" -trimpath \
      -o /out/npan-server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-s -w" -trimpath \
      -o /out/npan-cli ./cmd/cli

# Stage 2: Runtime
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S npan && adduser -S npan -G npan
COPY --from=builder /out/npan-server /usr/local/bin/npan-server
COPY --from=builder /out/npan-cli /usr/local/bin/npan-cli
COPY web/ /app/web/
RUN mkdir -p /app/data && chown -R npan:npan /app
VOLUME ["/app/data"]
WORKDIR /app
USER npan
EXPOSE 1323
HEALTHCHECK --interval=15s --timeout=3s --retries=3 \
    CMD wget -q -O /dev/null http://localhost:1323/healthz || exit 1
ENTRYPOINT ["npan-server"]
```

## 8. 优雅停机

```go
// cmd/server/main.go

func main() {
  // ...初始化...

  ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
  defer stop()

  go func() {
    if err := server.Start(cfg.ServerAddr); err != nil && err != http.ErrServerClosed {
      logger.Error("服务启动失败", "error", err)
      os.Exit(1)
    }
  }()

  <-ctx.Done()
  logger.Info("收到停机信号，开始优雅关闭...")

  shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
  defer cancel()

  if err := server.Shutdown(shutdownCtx); err != nil {
    logger.Error("优雅关闭失败", "error", err)
  }
}
```

## 9. 新增/修改文件清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `internal/config/validate.go` | 新建 | 配置启动校验 |
| `internal/httpx/middleware_auth.go` | 新建 | 认证中间件 |
| `internal/httpx/middleware_security.go` | 新建 | 安全头中间件 |
| `internal/httpx/middleware_ratelimit.go` | 新建 | 速率限制中间件 |
| `internal/httpx/errors.go` | 新建 | 统一错误响应 |
| `internal/httpx/dto.go` | 新建 | 响应 DTO |
| `internal/httpx/validation.go` | 新建 | 输入验证函数 |
| `internal/httpx/server.go` | 修改 | 路由重设计 + 中间件注册 |
| `internal/httpx/handlers.go` | 修改 | 移除 requireAPIAccess + 错误脱敏 + 输入校验 |
| `internal/search/meili_index.go` | 修改 | filter 注入防御 |
| `internal/search/query_service.go` | 修改 | pageSize 上限兜底 |
| `internal/npan/client.go` | 修改 | io.LimitReader |
| `cmd/server/main.go` | 修改 | 配置校验 + 优雅停机 |
| `web/demo/index.html` → `web/app/index.html` | 重命名+修改 | 生产化前端 |
| `Dockerfile` | 新建 | 多阶段构建 |
| `docker-compose.yml` | 修改 | 添加 npan 服务 |
