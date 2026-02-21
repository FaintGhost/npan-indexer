# Task 001: 编写错误响应测试 (Red)

**depends-on**: (无，独立任务)

## Description

在 `internal/httpx/errors_test.go` 中编写所有错误响应相关的测试，覆盖 5 个测试函数。测试需在实现代码存在之前先编写（TDD Red 阶段），此时测试预期编译失败或运行失败。

## Execution Context

**Task Number**: 001 of 002
**Phase**: Foundation
**Prerequisites**: 了解 Echo v5 的 `*echo.Context`、`echo.HTTPError`、`echo.HeaderXRequestID` 的用法；`newTestContext` 在 `handlers_test.go` 中已存在，但 `errors_test.go` 需自行创建测试 context。

## BDD Scenario Reference

**Scenarios** (内联于本计划，无单独 bdd-specs.md):

- Scenario 1: writeErrorResponse 返回正确的 JSON 结构
- Scenario 2: writeErrorResponse 在请求头含 X-Request-Id 时，响应中包含 request_id
- Scenario 3: customHTTPErrorHandler 接收 echo.HTTPError，转换为 ErrorResponse
- Scenario 4: customHTTPErrorHandler 接收普通 error，返回 500 + "服务器内部错误"
- Scenario 5: ErrorResponse 不包含内部信息字段（stack/trace/debug）

## Files to Create

- Create: `internal/httpx/errors_test.go`

## Steps

### Step 1: 创建 errors_test.go 并声明 package

文件头部使用 `package httpx`，导入 `encoding/json`、`errors`、`net/http`、`net/http/httptest`、`strings`、`testing`、`github.com/labstack/echo/v5`。

### Step 2: 创建本文件内的辅助函数 newEchoTestContext

与 `handlers_test.go` 中的 `newTestContext` 功能相同，但在本文件中独立定义（命名为 `newEchoTestContext` 以避免重复声明），接受 method、target 字符串，返回 `(*echo.Context, *httptest.ResponseRecorder)`，以便测试中检查响应体。

> 注意：需要同时返回 `*httptest.ResponseRecorder` 才能读取响应体，因为 `echo.Context` 的 `Response()` 返回的是 `*echo.Response`，而底层 writer 是 `*httptest.ResponseRecorder`。

构造方式：
1. 创建 `echo.New()`
2. 创建 `httptest.NewRequest(method, target, nil)`
3. 创建 `httptest.NewRecorder()`
4. 通过 `e.NewContext(req, rec)` 创建 context
5. 返回 context 和 rec

### Step 3: 实现 TestWriteErrorResponse_ReturnsJSON (Scenario 1)

- Given: 一个标准 GET 请求 context（无特殊头部）
- When: 调用 `writeErrorResponse(c, 400, "BAD_REQUEST", "缺少参数")`
- Then:
  - HTTP 状态码为 400
  - 响应 Content-Type 含 `application/json`
  - 反序列化 body 为 `ErrorResponse`，验证 `Code == "BAD_REQUEST"`，`Message == "缺少参数"`

### Step 4: 实现 TestWriteErrorResponse_IncludesRequestID (Scenario 2)

- Given: 请求头设置 `echo.HeaderXRequestID = "test-req-123"`
- When: 调用 `writeErrorResponse(c, 401, "UNAUTHORIZED", "未授权")`
- Then:
  - 反序列化 body，验证 `RequestID == "test-req-123"`

### Step 5: 实现 TestCustomHTTPErrorHandler_EchoError (Scenario 3)

- Given: 创建 `&echo.HTTPError{Code: 404, Message: "not found"}`
- When: 调用 `customHTTPErrorHandler(c, echoErr)`
- Then:
  - HTTP 状态码为 404
  - 反序列化 body 为 `ErrorResponse`，验证 `Code == "NOT_FOUND"`，`Message` 不为空

### Step 6: 实现 TestCustomHTTPErrorHandler_GenericError (Scenario 4)

- Given: 创建 `errors.New("some internal error")`
- When: 调用 `customHTTPErrorHandler(c, genericErr)`
- Then:
  - HTTP 状态码为 500
  - 反序列化 body，验证 `Code == "INTERNAL_ERROR"`，`Message == "服务器内部错误"`

### Step 7: 实现 TestErrorResponse_NoInternalInfo (Scenario 5)

- Given: 调用 `writeErrorResponse(c, 500, "INTERNAL_ERROR", "服务器内部错误")`
- When: 读取原始响应 body（字符串）
- Then:
  - body 中不含子串 `"stack"`、`"trace"`、`"debug"`

### Step 8: 验证测试编译失败（Red 阶段）

在 `errors.go` 不存在的情况下运行：

```bash
go test ./internal/httpx/ -run "TestWriteErrorResponse|TestCustomHTTPErrorHandler|TestErrorResponse" -v 2>&1 | head -20
```

预期：编译错误，提示 `writeErrorResponse`、`customHTTPErrorHandler`、`ErrorResponse` 未定义。

## Verification Commands

```bash
# Red 阶段验证（应该失败/编译错误）
go test ./internal/httpx/ -run "TestWriteErrorResponse|TestCustomHTTPErrorHandler|TestErrorResponse" -v
```

## Success Criteria

- `errors_test.go` 文件创建完毕，语法正确
- 在没有 `errors.go` 实现的情况下，`go test` 报告编译错误（符合 TDD Red 阶段预期）
- 5 个测试函数名称符合要求
