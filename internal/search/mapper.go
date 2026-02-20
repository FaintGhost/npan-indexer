package search

import "npan/internal/models"

func toSafeNumber(value int64, fallback int64) int64 {
	if value == 0 {
		return fallback
	}
	return value
}

func MapFolderToIndexDoc(folder models.NpanFolder, pathText string) models.IndexDocument {
	return models.IndexDocument{
		DocID:      "folder_" + formatInt(folder.ID),
		SourceID:   folder.ID,
		Type:       models.ItemTypeFolder,
		Name:       folder.Name,
		PathText:   pathText,
		ParentID:   folder.ParentID,
		ModifiedAt: toSafeNumber(folder.ModifiedAt, 0),
		CreatedAt:  0,
		Size:       0,
		SHA1:       "",
		InTrash:    folder.InTrash,
		IsDeleted:  folder.IsDeleted,
	}
}

func MapFileToIndexDoc(file models.NpanFile, pathText string) models.IndexDocument {
	return models.IndexDocument{
		DocID:      "file_" + formatInt(file.ID),
		SourceID:   file.ID,
		Type:       models.ItemTypeFile,
		Name:       file.Name,
		PathText:   pathText,
		ParentID:   file.ParentID,
		ModifiedAt: toSafeNumber(file.ModifiedAt, 0),
		CreatedAt:  toSafeNumber(file.CreatedAt, 0),
		Size:       toSafeNumber(file.Size, 0),
		SHA1:       file.SHA1,
		InTrash:    file.InTrash,
		IsDeleted:  file.IsDeleted,
	}
}

func formatInt(value int64) string {
	if value == 0 {
		return "0"
	}

	negative := value < 0
	if negative {
		value = -value
	}

	digits := make([]byte, 0, 20)
	for value > 0 {
		digits = append(digits, byte('0'+(value%10)))
		value /= 10
	}

	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}

	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}
