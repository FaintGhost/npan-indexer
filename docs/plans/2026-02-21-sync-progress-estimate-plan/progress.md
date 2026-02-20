# 进度记录

## 2026-02-21

- 已完成：创建“sync-full 估算进度增强”计划与 BDD 规格。
- 已完成：新增 `NpanFolder.item_count` 映射，并在部门根目录发现阶段生成 `estimatedTotalDocs=item_count+1`。
- 已完成：`SyncProgressState.RootProgress` 增加 `estimatedTotalDocs`，支持断点续跑恢复时刷新估算值。
- 已完成：`sync-full` human 输出增加估算维度 `est=xx.x%(docs=a/b roots=c/d)`，无估算时回退 `est=n/a`。
- 已完成：新增测试 `internal/service/sync_manager_estimate_test.go` 与 `internal/cli/root_progress_test.go`（估算场景）。

## 验证结果

- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/cli ./internal/service -run Estimate -count=1` 通过。
- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/cli ./internal/service ./internal/npan -count=1` 通过。
- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./... -count=1` 通过。
- `GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go run ./cmd/cli sync-full --help` 通过。
