# Task 046: 全量验收测试与代码质量检查

**depends-on**: task-045

## Description

运行全量测试套件、类型检查、lint 检查，确保所有验收标准满足。

## Execution Context

**Task Number**: 046 of 046
**Phase**: Final Verification
**Prerequisites**: 所有功能已实现，构建集成已完成

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: 所有场景的最终验证

## Files to Modify/Create

- (无新增文件，仅运行验证)

## Steps

### Step 1: Run full test suite

- `npx vitest run` — 所有单元测试和集成测试通过

### Step 2: Run TypeScript type check

- `npx tsc --noEmit` — strict mode 零错误

### Step 3: Run linter

- `npx oxlint` — 无错误

### Step 4: Run build

- `npm run build` — 构建成功
- 验证 gzip 后总体积 < 200KB（不含字体）

### Step 5: Manual smoke test checklist

- [ ] `/app` 显示搜索页 Hero 模式
- [ ] 输入关键词 → debounce 后搜索 → 显示结果
- [ ] 搜索触发 Hero → Docked 过渡（View Transition）
- [ ] 清空搜索 → Docked → Hero 过渡
- [ ] 滚动到底部 → 无限加载下一页
- [ ] 点击下载按钮 → 获取链接 → 新标签页打开
- [ ] Cmd/Ctrl+K → 聚焦搜索框
- [ ] `/app/admin` → API Key 输入对话框
- [ ] 输入有效 Key → 管理面板显示
- [ ] 启动同步 → 进度轮询 → 取消同步
- [ ] 刷新 `/app/admin` → SPA fallback 正常
- [ ] `/app?query=MX40` → 搜索框预填充并自动搜索

### Step 6: Go backend integration

- `go build ./cmd/server/` — 成功
- `go test ./...` — 后端测试仍通过（未改动后端逻辑）

## Verification Commands

```bash
cd cli
npx vitest run
npx tsc --noEmit
npx oxlint .
npm run build

cd /root/workspace/npan
go build ./cmd/server/
go test ./...
```

## Success Criteria

- 所有测试通过（vitest + go test）
- TypeScript strict 零错误
- oxlint 无错误
- 构建成功，产物可被 Go embed
- 手动冒烟测试全部通过
