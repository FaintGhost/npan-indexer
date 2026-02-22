# Task 006: Admin 同步 E2E 测试

**depends-on**: Task 004

## Objective

编写管理后台同步控制流程的 E2E 测试，覆盖 BDD specs Feature 3 中的同步相关场景。

## BDD Scenarios Covered

- Scenario: 显示同步模式选择器
- Scenario: 选择全量模式
- Scenario: 启动同步发送正确请求
- Scenario: 同步运行中显示进度
- Scenario: 取消同步需要确认
- Scenario: 取消确认框点击取消不发请求

## Files to Create/Modify

| File | Action |
|------|--------|
| `web/e2e/tests/admin.spec.ts` | 修改：添加同步测试 describe block |

## Steps

### 1. 在 admin.spec.ts 中添加同步测试 describe

同步测试需要已认证状态，使用 `authenticatedPage` fixture。

### 2. 编写「显示同步模式选择器」测试

- 使用已认证页面访问 `/admin/`
- 断言显示三个模式按钮：自适应、全量、增量
- 断言 "自适应" 为默认选中状态（通过 aria-pressed 或样式类判断）

### 3. 编写「选择全量模式」测试

- 点击 "全量" 按钮
- 断言 "全量" 按钮为选中状态
- 断言其他按钮为未选中状态

### 4. 编写「启动同步发送正确请求」测试

- 选择 "全量" 模式
- 使用 `page.waitForResponse()` 监听 POST `/api/v1/admin/sync`
- 点击 "启动同步" 按钮
- 验证请求头包含 `X-API-Key`
- 验证请求体包含 `mode: "full"`
- 断言显示成功提示

### 5. 编写「同步运行中显示进度」测试

- 启动同步（POST 返回 202）
- 断言页面开始轮询 GET `/api/v1/admin/sync`
- 断言显示进度信息
- 断言显示取消按钮

**注意**：由于 CI 中 NPA_TOKEN 为 dummy，同步会很快失败。测试应验证同步启动后的 UI 状态变化，而非等待同步完成。

### 6. 编写「取消同步需要确认」测试

- 启动同步
- 使用 `page.on('dialog')` 监听确认对话框
- 点击 "取消同步" 按钮
- 断言弹出确认对话框
- 接受对话框
- 断言发送 DELETE `/api/v1/admin/sync` 请求

### 7. 编写「取消确认框点击取消不发请求」测试

- 启动同步
- 监听 dialog 事件
- 点击 "取消同步"
- 在对话框中点击 "取消"（dismiss）
- 断言不发送 DELETE 请求
- 断言同步继续运行

## Verification

```bash
cd web && bunx playwright test admin.spec.ts --list
# 应列出认证 + 同步测试，同步部分约 6 个用例
```
