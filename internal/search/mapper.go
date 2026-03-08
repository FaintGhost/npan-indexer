package search

import (
	"regexp"
	"strings"

	"npan/internal/models"
)

// versionPrefixRe matches V/v immediately followed by digit.digit pattern,
// where V is either at word boundary or preceded by a non-ASCII character.
// Examples that match: "V1.5.0", "_V4.9.4.0", " v3.2.1"
// Examples that don't match: "VIVO", "V2", "固件包V2"
var versionPrefixRe = regexp.MustCompile(`([^a-zA-Z]|^)[Vv](\d+\.\d+)`)

var fileCategoryDocExtensions = map[string]bool{
	"pdf": true, "doc": true, "docx": true, "xls": true, "xlsx": true, "ppt": true, "pptx": true,
	"txt": true, "md": true, "markdown": true, "rtf": true, "odt": true, "ods": true, "odp": true,
	"csv": true, "tsv": true, "epub": true,
}

var fileCategoryImageExtensions = map[string]bool{
	"jpg": true, "jpeg": true, "png": true, "gif": true, "webp": true, "bmp": true, "svg": true,
	"tif": true, "tiff": true, "heic": true, "heif": true, "avif": true, "ico": true,
}

var fileCategoryVideoExtensions = map[string]bool{
	"mp4": true, "mkv": true, "mov": true, "avi": true, "wmv": true, "flv": true, "webm": true,
	"m4v": true, "mpg": true, "mpeg": true, "ts": true, "rmvb": true,
}

var fileCategoryMultiPartArchiveExtensions = []string{"tar.gz", "tar.bz2", "tar.xz"}

var fileCategoryArchiveExtensions = map[string]bool{
	"zip": true, "rar": true, "7z": true, "tar": true, "gz": true, "tgz": true, "bz2": true, "xz": true, "zst": true,
	"tar.gz": true, "tar.bz2": true, "tar.xz": true,
}

// ExtractNameExt returns the lowercase file extension if it is a known
// extension (present in knownExtensions), otherwise returns "".
func ExtractNameExt(name string) string {
	if name == "" {
		return ""
	}
	dotIdx := strings.LastIndex(name, ".")
	if dotIdx < 0 || dotIdx == len(name)-1 {
		return ""
	}
	ext := strings.ToLower(name[dotIdx+1:])
	if knownExtensions[ext] {
		return ext
	}
	return ""
}

// ExtractNameBase returns the file name with the known extension removed
// and any V/v prefix before version numbers stripped.
func ExtractNameBase(name string) string {
	if name == "" {
		return ""
	}

	base := name

	// Remove known extension (including the dot).
	ext := ExtractNameExt(name)
	if ext != "" {
		// Remove the last ".ext" from the name.
		base = name[:len(name)-len(ext)-1]
	}

	// Remove V/v prefix before version numbers (e.g., V1.5.0 -> 1.5.0).
	base = versionPrefixRe.ReplaceAllStringFunc(base, func(match string) string {
		// Find the position of V/v in the match.
		vIdx := strings.IndexAny(match, "Vv")
		prefix := match[:vIdx]
		rest := match[vIdx+1:] // everything after V/v
		return prefix + rest
	})

	return base
}

func categorizeFileName(name string) models.FileCategory {
	lowerName := strings.TrimSpace(strings.ToLower(name))
	if lowerName == "" {
		return models.FileCategoryOther
	}

	for _, ext := range fileCategoryMultiPartArchiveExtensions {
		if strings.HasSuffix(lowerName, "."+ext) {
			return models.FileCategoryArchive
		}
	}

	dotIdx := strings.LastIndex(lowerName, ".")
	if dotIdx <= 0 || dotIdx == len(lowerName)-1 {
		return models.FileCategoryOther
	}

	ext := lowerName[dotIdx+1:]
	switch {
	case fileCategoryDocExtensions[ext]:
		return models.FileCategoryDoc
	case fileCategoryImageExtensions[ext]:
		return models.FileCategoryImage
	case fileCategoryVideoExtensions[ext]:
		return models.FileCategoryVideo
	case fileCategoryArchiveExtensions[ext]:
		return models.FileCategoryArchive
	default:
		return models.FileCategoryOther
	}
}

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
		NameBase:   ExtractNameBase(folder.Name),
		NameExt:    "",
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
		DocID:        "file_" + formatInt(file.ID),
		SourceID:     file.ID,
		Type:         models.ItemTypeFile,
		Name:         file.Name,
		NameBase:     ExtractNameBase(file.Name),
		NameExt:      ExtractNameExt(file.Name),
		FileCategory: categorizeFileName(file.Name),
		PathText:     pathText,
		ParentID:     file.ParentID,
		ModifiedAt:   toSafeNumber(file.ModifiedAt, 0),
		CreatedAt:    toSafeNumber(file.CreatedAt, 0),
		Size:         toSafeNumber(file.Size, 0),
		SHA1:         file.SHA1,
		InTrash:      file.InTrash,
		IsDeleted:    file.IsDeleted,
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
