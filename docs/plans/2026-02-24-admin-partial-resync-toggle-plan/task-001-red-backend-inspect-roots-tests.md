# Task 001: RED backend inspect roots API tests

## Description

先为“拉取目录详情”接口建立失败测试，覆盖批量请求、部分成功与错误项返回，确保后续实现有明确行为边界。

## Execution Context

**Task Number**: 001 of 011  
**Phase**: Testing (Red)  
**Prerequisites**: 无

## BDD Scenario Reference

**Spec**: `../2026-02-24-admin-partial-resync-toggle-design/bdd-specs.md`  
**Scenario**: `批量拉取目录详情部分成功`

## Files to Modify/Create

- Modify: `internal/httpx/handlers_test.go`
- Modify: `internal/httpx/server_routes_test.go`（若需新增路由覆盖）

## Steps

### Step 1: Verify Scenario

- 确认 BDD 中“批量拉取目录详情部分成功”场景存在且语义明确。

### Step 2: Implement Test (Red)

- 为新接口编写 handler 层测试：
  - 输入多个 folder id
  - 模拟上游 API 对部分 id 返回成功、部分失败
  - 断言响应包含 `items` 与 `errors`
- 使用 test doubles 隔离外部依赖：
  - 使用 fake `npan.API`，禁止真实网络调用。

### Step 3: Verify Failure

- 运行新增用例，确认在未实现接口前失败，且失败原因是断言行为不满足（非编译错误）。

## Verification Commands

```bash
go test ./internal/httpx -run InspectRoots -count=1
```

## Success Criteria

- 测试稳定失败（Red）。
- 失败信息能直接指向“缺少 inspect roots 行为”。
