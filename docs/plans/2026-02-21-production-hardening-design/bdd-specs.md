# BDD 行为规范

## 1. API Key 认证中间件

```gherkin
Feature: API Key 认证中间件
  管理端点需要有效的 API Key；
  嵌入式前端端点（/api/v1/app/*）无需认证。

  Background:
    Given 服务配置 AdminAPIKey 为 "test-admin-key-32chars-minimum!!"

  Scenario: 未携带 API Key 访问管理端点返回 401
    When 客户端发送 GET "/api/v1/search/local?q=test" 不携带任何认证头
    Then 响应状态码应为 401
    And 响应体 JSON 包含 "code" 字段值为 "UNAUTHORIZED"
    And 响应体 JSON 不包含 "stack"、"path"、"config" 等内部信息

  Scenario: 携带错误 API Key 访问管理端点返回 401
    When 客户端发送 GET "/api/v1/search/local?q=test" 并设置 Header "X-API-Key" 为 "wrong-key"
    Then 响应状态码应为 401
    And 响应耗时与正确 Key 请求无显著差异（常量时间比较）

  Scenario: 通过 X-API-Key Header 认证成功
    When 客户端发送 GET "/api/v1/search/local?q=test" 并设置 Header "X-API-Key" 为 "test-admin-key-32chars-minimum!!"
    Then 响应状态码应为 200

  Scenario: 通过 Bearer Token 认证成功
    When 客户端发送 GET "/api/v1/search/local?q=test" 并设置 Header "Authorization" 为 "Bearer test-admin-key-32chars-minimum!!"
    Then 响应状态码应为 200

  Scenario: AdminAPIKey 为空时服务拒绝启动
    Given 服务配置 AdminAPIKey 为 ""
    When 服务尝试启动
    Then 服务应输出致命错误 "NPA_ADMIN_API_KEY 不能为空" 并退出

  Scenario Outline: 管理端点必须经过认证
    When 客户端发送 <method> "<path>" 不携带任何认证头
    Then 响应状态码应为 401

    Examples:
      | method | path                           |
      | GET    | /api/v1/search/remote?q=test   |
      | GET    | /api/v1/download-url?file_id=1 |
      | GET    | /api/v1/search/local?q=test    |
      | POST   | /api/v1/token                  |

  Scenario Outline: 管理端点（admin 组）必须经过认证
    When 客户端发送 <method> "<path>" 不携带任何认证头
    Then 响应状态码应为 401

    Examples:
      | method | path                                |
      | POST   | /api/v1/admin/sync/full             |
      | GET    | /api/v1/admin/sync/full/progress    |
      | POST   | /api/v1/admin/sync/full/cancel      |

  Scenario Outline: 公开端点不需要认证
    When 客户端发送 <method> "<path>" 不携带任何认证头
    Then 响应状态码不应为 401

    Examples:
      | method | path                                   |
      | GET    | /healthz                               |
      | GET    | /readyz                                |
      | GET    | /app                                   |
      | GET    | /api/v1/app/search?q=test              |
      | GET    | /api/v1/app/download-url?file_id=1     |
```

## 2. 速率限制

```gherkin
Feature: 速率限制
  防止暴力破解和资源滥用。

  Scenario: 单个 IP 超过请求速率限制返回 429
    Given 速率限制配置为每 IP 每秒 20 次请求
    When 同一 IP 在 1 秒内发送 25 次 GET "/api/v1/app/search?q=test"
    Then 前 20 次请求响应状态码应为 200
    And 后续请求响应状态码应为 429
    And 429 响应体 JSON 包含 "code" 字段值为 "RATE_LIMITED"
    And 429 响应包含 Header "Retry-After"

  Scenario: 不同 IP 各自独立计数
    Given 速率限制配置为每 IP 每秒 20 次请求
    When IP-A 在 1 秒内发送 15 次请求
    And IP-B 在 1 秒内发送 15 次请求
    Then IP-A 和 IP-B 的所有请求均返回 200

  Scenario: 速率限制窗口过后恢复正常
    Given 速率限制配置为每 IP 每秒 20 次请求
    When 同一 IP 在 1 秒内发送 25 次请求触发限流
    And 等待 1 秒
    And 同一 IP 再发送 1 次请求
    Then 该请求响应状态码应为 200
```

## 3. 输入验证

```gherkin
Feature: 输入验证
  所有用户输入经过严格验证。

  Scenario: pageSize 超过上限返回 400
    When 客户端发送 GET "/api/v1/search/local?q=test&page_size=1001"
    Then 响应状态码应为 400
    And 响应体 JSON "code" 字段值为 "BAD_REQUEST"

  Scenario: pageSize 为 0 或负数返回 400
    When 客户端发送 GET "/api/v1/search/local?q=test&page_size=-1"
    Then 响应状态码应为 400

  Scenario: pageSize 在有效范围内正常返回
    When 客户端发送 GET "/api/v1/search/local?q=test&page_size=50"
    Then 响应状态码应为 200

  Scenario Outline: type 参数只接受白名单值
    When 客户端发送 GET "/api/v1/search/local?q=test&type=<type_value>"
    Then 响应状态码应为 <expected_status>

    Examples:
      | type_value                     | expected_status |
      | all                            | 200             |
      | file                           | 200             |
      | folder                         | 200             |
      | file OR is_deleted = true      | 400             |
      | ' OR 1=1 --                    | 400             |

  Scenario Outline: checkpoint_template 路径遍历攻击被拒绝
    When 客户端发送 POST "/api/v1/admin/sync/full" 携带 "checkpoint_template" 为 "<path>"
    Then 响应状态码应为 400

    Examples:
      | path                           |
      | ../../../etc/passwd            |
      | /etc/shadow                    |
      | data/checkpoints/../../secrets |

  Scenario: 超大请求体返回 413
    When 客户端发送 POST "/api/v1/admin/sync/full" 携带 2MB 的请求体
    Then 响应状态码应为 413
```

## 4. 错误响应格式

```gherkin
Feature: 错误响应格式
  所有错误响应使用统一格式，不泄露内部实现细节。

  Scenario: 错误响应使用统一 JSON 结构
    When 任意请求返回 4xx 或 5xx 状态码
    Then 响应体 JSON 包含 "code" 字段（字符串类型）
    And 响应体 JSON 包含 "message" 字段（字符串类型）
    And 响应体 JSON 不包含 "stack"、"trace"、"debug" 等字段

  Scenario: 500 错误不泄露堆栈
    When 服务内部出现 panic
    Then 响应状态码应为 500
    And 响应体 JSON "message" 字段值为 "服务器内部错误"
    And 响应体不包含文件路径、行号或 Go 堆栈信息

  Scenario: Meilisearch 错误不直接透传
    When Meilisearch 返回错误
    Then API 响应体 "message" 字段值为 "搜索服务暂不可用"
    And 响应体不包含 "meilisearch"、"meili" 等字样

  Scenario: Token 获取失败不泄露 Client Secret
    When Token 端点因凭据错误返回失败
    Then 响应体不包含 "client_secret" 的实际值
    And 响应体 "message" 字段值为 "认证失败，请检查凭据"

  Scenario: sync/full/progress 不泄露内部配置
    When 客户端请求 GET "/api/v1/admin/sync/full/progress"
    Then 响应体不包含 "meiliHost"、"meiliIndex"、"checkpointTemplate" 字段
```

## 5. 凭据管理

```gherkin
Feature: 凭据管理
  秘密信息不出现在响应、日志或版本控制中。

  Scenario: API 响应不包含服务端凭据
    When 客户端发送任意请求并收到响应
    Then 响应体不包含 NPA_CLIENT_SECRET 的值
    And 响应体不包含 MEILI_API_KEY 的值
    And 响应体不包含 NPA_ADMIN_API_KEY 的值

  Scenario: 日志不打印敏感字段
    When 服务处理请求并记录日志
    Then 日志输出不包含 "client_secret" 的实际值
    And 日志中凭据类字段显示为 "[REDACTED]"

  Scenario: .env 文件不在 git 中
    When 执行 "git ls-files" 命令
    Then 输出不包含 ".env"（排除 .env.example）
    And 输出不包含 ".env.meilisearch"（排除 .env.meilisearch.example）
```

## 6. 健康检查

```gherkin
Feature: 健康检查端点

  Scenario: healthz 始终返回 200
    When 客户端发送 GET "/healthz"
    Then 响应状态码应为 200
    And 响应体 JSON 包含 "status" 字段值为 "ok"

  Scenario: readyz 在 Meilisearch 可用时返回 200
    Given Meilisearch 服务正常运行
    When 客户端发送 GET "/readyz"
    Then 响应状态码应为 200
    And 响应体 JSON 包含 "status" 字段值为 "ready"

  Scenario: readyz 在 Meilisearch 不可达时返回 503
    Given Meilisearch 服务不可用
    When 客户端发送 GET "/readyz"
    Then 响应状态码应为 503
    And 响应体 JSON 包含 "status" 字段值为 "not_ready"
```
