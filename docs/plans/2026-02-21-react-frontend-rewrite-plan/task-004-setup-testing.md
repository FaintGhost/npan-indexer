# Task 004: 配置 Vitest + React Testing Library + MSW

**depends-on**: task-001

## Description

配置前端测试基础设施：Vitest 作为测试运行器，React Testing Library 用于组件测试，MSW 用于 API Mock。

## Execution Context

**Task Number**: 004 of 046
**Phase**: Setup
**Prerequisites**: Task 001 项目已初始化

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: (Infrastructure — 所有测试任务依赖此配置)

## Files to Modify/Create

- Modify: `cli/vite.config.ts` — 添加 Vitest 配置
- Create: `cli/vitest.config.ts`（或在 vite.config.ts 中配置 test）
- Create: `cli/src/tests/setup.ts` — 测试全局 setup（jsdom、afterEach cleanup）
- Create: `cli/src/tests/mocks/handlers.ts` — MSW handlers 骨架
- Create: `cli/src/tests/mocks/server.ts` — MSW setupServer
- Create: `cli/src/tests/sample.test.ts` — 验证测试基础设施可用的示例测试

## Steps

### Step 1: Install test dependencies

- Install: `vitest`, `@testing-library/react`, `@testing-library/jest-dom`, `@testing-library/user-event`, `jsdom`, `msw`

### Step 2: Configure Vitest

- 配置 `environment: 'jsdom'`
- 配置 `setupFiles: ['./src/tests/setup.ts']`
- 配置 `globals: true`（可选）
- 确保 path alias `@/` 在测试中也能解析

### Step 3: Create test setup

- 在 `setup.ts` 中导入 `@testing-library/jest-dom/vitest`
- 配置 `afterEach(() => cleanup())`
- 配置 MSW server `beforeAll(server.listen)`, `afterEach(server.resetHandlers)`, `afterAll(server.close)`

### Step 4: Create MSW handlers skeleton

- 创建基础 handlers 数组（搜索、下载、同步进度的默认 mock）
- 创建 `setupServer(...handlers)` 导出

### Step 5: Create sample test

- 创建一个简单的 React 组件渲染测试，验证整个工具链正常

### Step 6: Verify

- `npm run test` 或 `npx vitest run` 通过

## Verification Commands

```bash
cd cli && npx vitest run
```

## Success Criteria

- `vitest run` 执行示例测试并通过
- React Testing Library render/screen 可正常使用
- MSW server 可拦截 fetch 请求
- jest-dom matcher（如 `toBeInTheDocument`）可用
