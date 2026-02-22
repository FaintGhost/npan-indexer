# Task 008: 全链路验证

**depends-on**: Task 002, Task 003, Task 004, Task 005, Task 006, Task 007

## Objective

在 Docker Compose CI 环境中运行完整 E2E 测试套件，验证所有测试通过，确认 CI pipeline 正常工作。

## Files to Create/Modify

| File | Action |
|------|--------|
| (无新文件) | 仅验证和修复 |

## Steps

### 1. 本地 Docker Compose 全流程验证

```bash
# 启动服务
docker compose -f docker-compose.ci.yml up --build -d --wait --wait-timeout 120

# 运行冒烟测试确认服务健康
BASE_URL=http://localhost:11323 METRICS_URL=http://localhost:19091 ./tests/smoke/smoke_test.sh

# 运行 E2E 测试
docker compose -f docker-compose.ci.yml run --rm --profile e2e playwright

# 清理
docker compose -f docker-compose.ci.yml down --volumes
```

或使用 Makefile target：

```bash
make e2e-test
```

### 2. 验证检查项

- [ ] 所有 E2E 测试通过（0 failures）
- [ ] 搜索流程测试：初始状态、防抖搜索、点击搜索、Enter 搜索、空状态、清空恢复、无限滚动、快捷键、视图切换
- [ ] 下载流程测试：初始状态、成功状态、失败重试、缓存、并行下载
- [ ] Admin 认证测试：对话框显示、空 key 错误、错误 key 错误、正确 key 认证、刷新保持、返回搜索
- [ ] Admin 同步测试：模式选择、启动同步、同步进度、取消确认、取消拒绝
- [ ] 边界场景测试：特殊字符、长查询、竞态、网络错误、纯空格、导航、认证过期
- [ ] Playwright report 生成在 `web/playwright-report/`
- [ ] 失败时 screenshot 保存在 `web/test-results/`
- [ ] Docker 容器正确清理（无残留）

### 3. 修复发现的问题

如果验证过程中发现问题：
- 修复测试代码中的定位器或断言
- 修复 fixtures 中的播种/认证逻辑
- 修复 Docker Compose 配置
- 调整超时或等待策略

### 4. 确认 GitHub Actions 兼容性

- 验证 `docker-compose.ci.yml` 配置语法
- 验证 artifact upload 路径正确
- 确认 `if: ${{ !cancelled() }}` 条件正确

## Verification

```bash
# 最终验证：完整 e2e-test target
make e2e-test
# 应输出所有测试通过，无错误
```
