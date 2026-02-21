package httpx

import (
  "fmt"
  "path/filepath"
  "strings"
)

const maxPageSize int64 = 100

var allowedTypes = map[string]bool{
  "all": true, "file": true, "folder": true,
}

func validatePageSize(pageSize int64) error {
  if pageSize <= 0 {
    return fmt.Errorf("page_size 必须是正整数")
  }
  if pageSize > maxPageSize {
    return fmt.Errorf("page_size 不能超过 %d", maxPageSize)
  }
  return nil
}

func validateType(typeParam string) error {
  if typeParam == "" {
    return nil
  }
  if !allowedTypes[typeParam] {
    return fmt.Errorf("type 参数无效，允许值: all, file, folder")
  }
  return nil
}

func validateCheckpointTemplate(template string) error {
  if template == "" {
    return nil
  }
  cleaned := filepath.Clean(template)
  if filepath.IsAbs(cleaned) {
    return fmt.Errorf("检查点路径无效: 不允许绝对路径")
  }
  if strings.Contains(cleaned, "..") {
    return fmt.Errorf("检查点路径无效: 不允许路径遍历")
  }
  if !strings.HasPrefix(cleaned, "data"+string(filepath.Separator)+"checkpoints") &&
    !strings.HasPrefix(cleaned, "data/checkpoints") {
    return fmt.Errorf("检查点路径无效: 必须在 data/checkpoints 目录下")
  }
  return nil
}
