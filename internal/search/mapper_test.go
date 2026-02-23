package search

import "testing"

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
