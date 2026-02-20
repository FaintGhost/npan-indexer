package indexer

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"npan/internal/models"
	"npan/internal/search"
)

type UpdatedWindowFetcher func(ctx context.Context, start *int64, end *int64, pageID int64) (map[string]any, error)

type IncrementalFetchOptions struct {
	Since int64
	Until int64
	Retry models.RetryPolicyOptions
	Fetch UpdatedWindowFetcher
}

func FetchIncrementalChanges(ctx context.Context, opts IncrementalFetchOptions) ([]IncrementalInputItem, error) {
	if opts.Fetch == nil {
		return nil, fmt.Errorf("缺少 Fetch 函数")
	}

	start := opts.Since
	end := opts.Until

	pageID := int64(0)
	changesByID := map[string]IncrementalInputItem{}

	for {
		page, err := WithRetry(ctx, func() (map[string]any, error) {
			return opts.Fetch(ctx, &start, &end, pageID)
		}, opts.Retry)
		if err != nil {
			return nil, err
		}

		files := asMapList(page["files"])
		for _, row := range files {
			id := toInt64(row["id"], 0)
			if id <= 0 {
				continue
			}

			name := strings.TrimSpace(toString(row["name"]))
			parentID := extractParentID(row)
			inTrash := toBool(row["in_trash"])
			isDeleted := toBool(row["is_deleted"])

			file := models.NpanFile{
				ID:         id,
				Name:       name,
				ParentID:   parentID,
				Size:       toInt64(row["size"], 0),
				ModifiedAt: toInt64(row["modified_at"], 0),
				CreatedAt:  toInt64(row["created_at"], 0),
				SHA1:       strings.TrimSpace(toString(row["sha1"])),
				InTrash:    inTrash,
				IsDeleted:  isDeleted,
			}

			doc := search.MapFileToIndexDoc(file, resolvePathText(row, "file", id, name))
			changesByID[doc.DocID] = IncrementalInputItem{
				Doc:     doc,
				Deleted: inTrash || isDeleted,
			}
		}

		folders := asMapList(page["folders"])
		for _, row := range folders {
			id := toInt64(row["id"], 0)
			if id <= 0 {
				continue
			}

			name := strings.TrimSpace(toString(row["name"]))
			parentID := extractParentID(row)
			inTrash := toBool(row["in_trash"])
			isDeleted := toBool(row["is_deleted"])

			folder := models.NpanFolder{
				ID:         id,
				Name:       name,
				ParentID:   parentID,
				ModifiedAt: toInt64(row["modified_at"], 0),
				InTrash:    inTrash,
				IsDeleted:  isDeleted,
			}

			doc := search.MapFolderToIndexDoc(folder, resolvePathText(row, "folder", id, name))
			changesByID[doc.DocID] = IncrementalInputItem{
				Doc:     doc,
				Deleted: inTrash || isDeleted,
			}
		}

		pageCount := toInt64(page["page_count"], 1)
		if pageCount <= 0 {
			pageCount = 1
		}

		pageID++
		if pageID >= pageCount {
			break
		}
	}

	keys := make([]string, 0, len(changesByID))
	for key := range changesByID {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]IncrementalInputItem, 0, len(keys))
	for _, key := range keys {
		result = append(result, changesByID[key])
	}

	return result, nil
}

func asMapList(input any) []map[string]any {
	if input == nil {
		return nil
	}

	if typed, ok := input.([]map[string]any); ok {
		return typed
	}

	rows, ok := input.([]any)
	if !ok {
		return nil
	}

	result := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		mapped, ok := row.(map[string]any)
		if !ok {
			continue
		}
		result = append(result, mapped)
	}

	return result
}

func extractParentID(item map[string]any) int64 {
	parentID := toInt64(item["parent_id"], 0)
	if parentRaw, ok := item["parent"].(map[string]any); ok {
		return toInt64(parentRaw["id"], parentID)
	}
	return parentID
}

func resolvePathText(item map[string]any, kind string, sourceID int64, name string) string {
	if pathRaw, ok := item["path_text"].(string); ok {
		path := strings.TrimSpace(pathRaw)
		if path != "" {
			return path
		}
	}

	return fmt.Sprintf("%s/%d/%s", kind, sourceID, name)
}

func toInt64(input any, fallback int64) int64 {
	switch value := input.(type) {
	case float64:
		return int64(value)
	case int64:
		return value
	case int:
		return int64(value)
	case int32:
		return int64(value)
	case string:
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return fallback
		}
		sign := int64(1)
		if trimmed[0] == '-' {
			sign = -1
			trimmed = trimmed[1:]
		}
		if trimmed == "" {
			return fallback
		}

		result := int64(0)
		for i := 0; i < len(trimmed); i++ {
			ch := trimmed[i]
			if ch < '0' || ch > '9' {
				return fallback
			}
			result = result*10 + int64(ch-'0')
		}
		return sign * result
	default:
		return fallback
	}
}

func toBool(input any) bool {
	switch value := input.(type) {
	case bool:
		return value
	case string:
		normalized := strings.ToLower(strings.TrimSpace(value))
		return normalized == "true" || normalized == "1" || normalized == "yes"
	case float64:
		return value != 0
	case int:
		return value != 0
	case int64:
		return value != 0
	default:
		return false
	}
}

func toString(input any) string {
	if input == nil {
		return ""
	}
	return fmt.Sprintf("%v", input)
}
