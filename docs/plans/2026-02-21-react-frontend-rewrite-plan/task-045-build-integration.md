# Task 045: 构建产物集成（Vite build → Go embed.FS）

**depends-on**: task-042, task-044

## Description

配置 Vite 构建产物输出，调整 Go 服务端以支持 embed.FS 服务 SPA 静态资源和 SPA fallback 路由。

## Execution Context

**Task Number**: 045 of 046
**Phase**: Build Integration
**Prerequisites**: 所有前端功能已实现

## BDD Scenario Reference

**Spec**: `../2026-02-21-react-frontend-rewrite-design/bdd-specs.md`
**Scenario**: Feature 7 - 浏览器直接访问 /app/admin; Feature 7 - 管理页懒加载

## Files to Modify/Create

- Modify: `cli/vite.config.ts` — 确认 build.outDir 和 base 配置
- Modify: `cli/index.html` — 确保 SPA 入口正确
- Create/Modify: Go 服务端 embed 配置（`web/embed.go` 或 `internal/httpx/server.go`）

## Steps

### Step 1: Verify Vite build configuration

- `base: "/app/"` — SPA 基础路径
- `build.outDir: "../web/dist"` — 输出到 Go 可 embed 的目录
- `build.emptyOutDir: true`
- ManualChunks: vendor (react/react-dom), router (@tanstack/react-router)

### Step 2: Run build and verify output

- `npm run build`
- 验证 `web/dist/` 包含 index.html 和 assets/
- 验证 admin 路由有独立 chunk（代码分割生效）

### Step 3: Configure Go embed.FS

- 使用 `//go:embed dist/*` 嵌入 `web/dist/`
- 配置 `/app` 和 `/app/*` 路由的 SPA fallback
- 所有未匹配的 `/app/*` 路径返回 `index.html`
- 静态资源（js/css/images）直接服务

### Step 4: Verify end-to-end

- `go build` 成功
- 启动 Go 服务 → 访问 `/app` 返回前端页面
- 访问 `/app/admin` 返回前端页面（SPA fallback）
- 静态资源（JS/CSS）正确加载

## Verification Commands

```bash
cd cli && npm run build
ls -la ../web/dist/
ls -la ../web/dist/assets/
cd /root/workspace/npan && go build ./cmd/server/
```

## Success Criteria

- `npm run build` 成功输出到 `web/dist/`
- Admin 页面有独立 JS chunk
- Go 服务可以 embed 并服务前端资源
- SPA fallback 正确处理客户端路由
