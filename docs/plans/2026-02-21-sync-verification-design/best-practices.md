# Best Practices

## Error Handling

### Retry Policy

- **API 请求**（ListFolderChildren）: 已有 `WithRetry`，重试 429/5xx/网络错误
- **MeiliSearch 写入**（UpsertDocuments）: 新增 `WithRetryVoid` 包裹，复用相同的 `RetryPolicyOptions`
- **`isRetriable` 扩展**: 必须识别 `*meilisearch.Error` 类型，否则 MeiliSearch 的 429/5xx 会被当作不可重试

### Graceful Degradation

- Upsert 重试耗尽后：记录 SkippedFiles，**继续爬取下一页**（而非终止）
- 对账阶段 GetStats 失败：`Verification` 设为 nil，不影响 sync status 为 "done"
- 这确保了一个 MeiliSearch 短暂故障不会导致整个多小时的爬取作废

## Counter Accuracy

### FilesIndexed 修复原则

移动计数器到 Upsert 成功后是最小改动，保证：
- `FilesIndexed` = 成功写入 MeiliSearch 的文件数
- `FilesDiscovered` = 从 API 看到的文件总数
- `SkippedFiles` = 重试耗尽后跳过的文件数
- 不变量: `FilesDiscovered == FilesIndexed + SkippedFiles`（理论上）

### 为什么不用 TotalCount

API 返回的 `page.TotalCount` 语义不完全清晰（可能含回收站项目、可能包含文件夹），且只有第一页响应时才可靠。改为**逐页累加 `len(page.Files)`** 更精确，因为这就是我们尝试索引的精确文件集。

## Performance Considerations

- `MeiliIndex.DocumentCount()` 只是一个 GET 请求，无性能影响
- `WithRetryVoid` 是纯函数包装，零分配开销
- 新增计数器是 `int64` 累加，可忽略
- 失败跳过避免了"一个坏批次终止整个爬取"的问题，实际上提升了整体吞吐

## Security

- 无新的外部输入点
- 对账数据仅通过已有的 admin API 端点暴露（受 APIKeyAuth 保护）
- 不引入新的依赖

## Testing

- 所有新逻辑必须有单元测试
- `isRetriable` 扩展需要针对 `*meilisearch.Error` 的具体 case 测试
- `full_crawl.go` 改动需验证计数器在成功/失败场景下的准确性
- 前端 schema 变更需验证向后兼容（旧数据缺少新字段时不报错）
