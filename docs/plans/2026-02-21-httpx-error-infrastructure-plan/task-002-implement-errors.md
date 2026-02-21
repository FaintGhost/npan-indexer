# Task 002: 实现 errors.go (Green)

**depends-on**: task-001-write-error-response-tests

## Description

创建 `internal/httpx/errors.go`，实现 `ErrorResponse` struct、错误码常量、`writeErrorResponse` 函数和 `customHTTPErrorHandler` 全局错误处理器，使 task-001 中编写的所有测试通过。

## Execution Context

**Task Number**: 002 of 002
**Phase**: Core Features
**Prerequisites**: task-001 完成，`errors_test.go` 已存在且测试处于 Red 状态（编译失败）。

## BDD Scenario Reference

**Scenarios** (与 task-001 一致):

- Scenario 1: writeErrorResponse 返回正确的 JSON 结构
- Scenario 2: writeErrorResponse 响应包含 request_id
- Scenario 3: customHTTPErrorHandler 处理 echo.HTTPError
- Scenario 4: customHTTPErrorHandler 处理普通 error，返回 500
- Scenario 5: ErrorResponse 不含内部调试字段

## Files to Create

- Create: `internal/httpx/errors.go`

## Steps

### Step 1: 创建 errors.go，声明 package 和导入

- Package: `httpx`
- 导入: `encoding/json`（可选，通常通过 echo 的 `c.JSON` 即可）、`errors`、`fmt`（用于格式化错误信息）、`log/slog`、`net/http`、`github.com/labstack/echo/v5`

### Step 2: 定义 ErrorResponse struct

字段：
- `Code string` JSON tag `json:"code"`
- `Message string` JSON tag `json:"message"`
- `RequestID string` JSON tag `json:"request_id,omitempty"`

### Step 3: 定义错误码常量

```
ErrCodeUnauthorized  = "UNAUTHORIZED"
ErrCodeBadRequest    = "BAD_REQUEST"
ErrCodeNotFound      = "NOT_FOUND"
ErrCodeConflict      = "CONFLICT"
ErrCodeRateLimited   = "RATE_LIMITED"
ErrCodeInternalError = "INTERNAL_ERROR"
```

### Step 4: 实现 writeErrorResponse 函数

签名：`func writeErrorResponse(c *echo.Context, status int, code string, message string) error`

逻辑：
1. 从请求头提取 Request ID：`c.Request().Header.Get(echo.HeaderXRequestID)`
2. 构建 `ErrorResponse{Code: code, Message: message, RequestID: requestID}`
3. 调用 `c.JSON(status, resp)` 并返回其结果

### Step 5: 实现 customHTTPErrorHandler 函数

签名：`func customHTTPErrorHandler(c *echo.Context, err error)`

逻辑：
1. 声明 `status int` 和 `code string` 和 `message string`
2. 尝试用 `errors.As(err, &he)` 将 err 转为 `*echo.HTTPError`：
   - 若成功：`status = he.Code`，根据 status 映射到对应 `ErrCode*` 常量：
     - 400 → `ErrCodeBadRequest`
     - 401 → `ErrCodeUnauthorized`
     - 404 → `ErrCodeNotFound`
     - 409 → `ErrCodeConflict`
     - 429 → `ErrCodeRateLimited`
     - 其他 → `ErrCodeInternalError`
   - `message` 取 `fmt.Sprintf("%v", he.Message)`（避免泄露内部细节，仅展示 echo 定义的消息）
   - 若 status >= 500：用 slog 记录 error 级别日志，包含 err 信息
3. 若转换失败（普通 error）：
   - `status = 500`，`code = ErrCodeInternalError`，`message = "服务器内部错误"`
   - 用 `slog.Error` 记录完整错误，包含 err 信息
4. 调用 `writeErrorResponse(c, status, code, message)`，忽略其返回值（handler 无返回值）

### Step 6: 验证所有测试通过 (Green 阶段)

```bash
go test ./internal/httpx/ -run "TestWriteErrorResponse|TestCustomHTTPErrorHandler|TestErrorResponse" -v
```

预期：全部 5 个测试 PASS。

### Step 7: 运行完整 httpx 包测试，确保无回归

```bash
go test ./internal/httpx/ -v
```

预期：所有测试通过。

## Verification Commands

```bash
# Green 阶段验证（应该全部通过）
go test ./internal/httpx/ -run "TestWriteErrorResponse|TestCustomHTTPErrorHandler|TestErrorResponse" -v

# 完整包测试（确保无回归）
go test ./internal/httpx/ -v
```

## Success Criteria

- 全部 5 个测试函数通过
- `go test ./internal/httpx/ -v` 无任何失败
- `ErrorResponse` 序列化不含 `stack`/`trace`/`debug` 字段
- `customHTTPErrorHandler` 对 500+ 错误有 slog 日志记录
