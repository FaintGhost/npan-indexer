# Best Practices

## 1. 根因修复优先原则

- 先修“状态一致性”再修“展示层”。
- `force_rebuild` 的语义应是“索引与断点双重重置”。
- 避免通过 UI 文案掩盖后端状态问题。

## 2. 使用官方 API 做统计时的约束

### Folder Info

- 接口：`GET /api/v2/folder/{id}/info`
- 返回包含 `item_count`、`name`，适合作为 root 级预估来源。
- 该预估应作为“校验参考”，不是强一致真值（受权限、回收站、筛选策略影响）。

### Folder Children

- 接口：`GET /api/v2/folder/{id}/children`
- 返回 `page_id/page_count/page_capacity/total_count`，用于爬取流程与分页排障。
- 全量流程仍应以 `page_count` 作为停止条件，不要自行猜测页数。

### Item Search

- 接口：`GET /api/v2/item/search`
- 支持 `search_in_folder`、`type`、`page_id`、`query_filter`、`updated_time_range`。
- 若后续需要更深的目录递归统计，可基于此接口扩展。

## 3. 降级与容错

- `GetFolderInfo` 失败不能阻断全量同步，只影响 estimate 展示。
- 强制重建前清理 checkpoint 失败应明确报错并终止（避免“半重置”）。
- 警告阈值应可配置（建议默认：差异比 > 5% 或绝对差 > 20）。

## 4. 兼容性策略

- 请求字段复用已有契约（`root_folder_ids/include_departments`），避免新增 API 破坏面。
- UI 新增输入必须兼容空值：空值时行为与当前版本一致。
- 进度结构尽量复用已存在字段（`estimatedTotalDocs`, `verification.warnings`）。

## 5. 可观测性建议

- 在日志中增加关键字段：
  - `mode`, `force_rebuild`, `resume_progress`
  - `root_id`, `checkpoint_file`, `checkpoint_cleared`
  - `estimated_total_docs`, `actual_docs`, `diff`
- 这样可直接定位“统计偏小”是断点问题、权限问题还是 API 返回问题。

## 6. 安全与最小权限

- 目录范围索引能力仅在 Admin API 下开放（现有 API Key 保护范围内）。
- 不在公开 `/app` 路径暴露目录统计管理操作。
- 不在前端持久化服务端凭据。

## References

- FangCloud OpenAPI v3 Wiki: https://open.fangcloud.com/wiki/v3/
- Folder Children: https://open.fangcloud.com/wiki/v3/api/folder/children.html
- Folder Info: https://open.fangcloud.com/wiki/v3/api/folder/info.html
- Item Search: https://open.fangcloud.com/wiki/v3/api/item/search.html
