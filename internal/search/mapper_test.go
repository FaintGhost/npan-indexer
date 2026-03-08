package search

import (
  "reflect"
  "strings"
  "testing"

  "npan/internal/models"
)

func TestExtractNameBase(t *testing.T) {
  tests := []struct {
    input string
    want  string
  }{
    {"MX6000_2000 Pro V1.5.0.zip", "MX6000_2000 Pro 1.5.0"},
    {"4.9.2.0支持IC.png", "4.9.2.0支持IC"},
    {"DATA_A5sPlus_V4.9.4.0.zip", "DATA_A5sPlus_4.9.4.0"},
    {"王晨周报.pdf", "王晨周报"},
    {"README", "README"},
    {"固件包V2", "固件包V2"},
    {"archive.tar.gz", "archive.tar"},
    {"firmware_v3.2.1.bin", "firmware_3.2.1"},
    {"", ""},
    {"test.unknownext", "test.unknownext"},
  }
  for _, tt := range tests {
    t.Run(tt.input, func(t *testing.T) {
      got := ExtractNameBase(tt.input)
      if got != tt.want {
        t.Errorf("ExtractNameBase(%q) = %q, want %q", tt.input, got, tt.want)
      }
    })
  }
}

func TestExtractNameExt(t *testing.T) {
  tests := []struct {
    input string
    want  string
  }{
    {"MX6000_2000 Pro V1.5.0.zip", "zip"},
    {"王晨周报.pdf", "pdf"},
    {"README", ""},
    {"固件包V2", ""},
    {"archive.tar.gz", "gz"},
    {"firmware_v3.2.1.bin", "bin"},
    {"", ""},
    {"test.unknownext", ""},
    {"photo.JPG", "jpg"},
  }
  for _, tt := range tests {
    t.Run(tt.input, func(t *testing.T) {
      got := ExtractNameExt(tt.input)
      if got != tt.want {
        t.Errorf("ExtractNameExt(%q) = %q, want %q", tt.input, got, tt.want)
      }
    })
  }
}

func TestMapFileToIndexDoc_MapsFileCategoryByExtension(t *testing.T) {
  tests := []struct {
    name string
    file models.NpanFile
    want string
  }{
    {
      name: "documents use doc category",
      file: models.NpanFile{ID: 101, Name: "季度汇报.PDF"},
      want: "doc",
    },
    {
      name: "images use image category",
      file: models.NpanFile{ID: 102, Name: "封面图.JPG"},
      want: "image",
    },
    {
      name: "videos use video category",
      file: models.NpanFile{ID: 103, Name: "演示视频.MP4"},
      want: "video",
    },
    {
      name: "archives use archive category",
      file: models.NpanFile{ID: 104, Name: "交付包.tar.gz"},
      want: "archive",
    },
    {
      name: "unknown extension falls back to other",
      file: models.NpanFile{ID: 105, Name: "README"},
      want: "other",
    },
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      doc := MapFileToIndexDoc(tt.file, "root/shared")
      category, ok := readIndexDocumentStringField(t, doc, "FileCategory")
      if !ok {
        t.Fatalf("IndexDocument must expose FileCategory field for %q", tt.file.Name)
      }
      if category != tt.want {
        t.Fatalf("MapFileToIndexDoc(%q) FileCategory = %q, want %q", tt.file.Name, category, tt.want)
      }
    })
  }
}

func readIndexDocumentStringField(t *testing.T, doc models.IndexDocument, fieldName string) (string, bool) {
  t.Helper()

  value := reflect.ValueOf(doc)
  field := value.FieldByName(fieldName)
  if !field.IsValid() {
    return "", false
  }
  if field.Kind() != reflect.String {
    t.Fatalf("IndexDocument.%s kind = %s, want string", fieldName, field.Kind())
  }

  return field.String(), true
}

func TestIndexDocument_FileCategoryJsonContract(t *testing.T) {
  docType := reflect.TypeOf(models.IndexDocument{})
  field, ok := docType.FieldByName("FileCategory")
  if !ok {
    t.Fatal("IndexDocument must declare FileCategory field")
  }

  if got := field.Type.Kind(); got != reflect.String {
    t.Fatalf("IndexDocument.FileCategory kind = %s, want string", got)
  }

  jsonTag := field.Tag.Get("json")
  if !strings.Contains(jsonTag, "file_category") {
    t.Fatalf("IndexDocument.FileCategory json tag = %q, want to contain %q", jsonTag, "file_category")
  }
}
