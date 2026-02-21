package httpx

import "testing"

// --- validatePageSize ---

func TestValidatePageSize_ExceedsMax_ReturnsError(t *testing.T) {
  err := validatePageSize(1001)
  if err == nil {
    t.Fatal("expected error for pageSize=1001")
  }
}

func TestValidatePageSize_Negative_ReturnsError(t *testing.T) {
  err := validatePageSize(-1)
  if err == nil {
    t.Fatal("expected error for pageSize=-1")
  }
}

func TestValidatePageSize_Zero_ReturnsError(t *testing.T) {
  err := validatePageSize(0)
  if err == nil {
    t.Fatal("expected error for pageSize=0")
  }
}

func TestValidatePageSize_ValidRange_NoError(t *testing.T) {
  err := validatePageSize(50)
  if err != nil {
    t.Fatalf("expected no error for pageSize=50, got: %v", err)
  }
}

func TestValidatePageSize_MaxBoundary_NoError(t *testing.T) {
  err := validatePageSize(100)
  if err != nil {
    t.Fatalf("expected no error for pageSize=100, got: %v", err)
  }
}

func TestValidatePageSize_OverMaxBoundary_ReturnsError(t *testing.T) {
  err := validatePageSize(101)
  if err == nil {
    t.Fatal("expected error for pageSize=101")
  }
}

// --- validateType ---

func TestValidateType_AllowedValues(t *testing.T) {
  cases := []struct {
    input string
  }{
    {"all"},
    {"file"},
    {"folder"},
  }
  for _, tc := range cases {
    err := validateType(tc.input)
    if err != nil {
      t.Errorf("expected no error for type=%q, got: %v", tc.input, err)
    }
  }
}

func TestValidateType_EmptyString_Allowed(t *testing.T) {
  err := validateType("")
  if err != nil {
    t.Fatalf("expected no error for empty type, got: %v", err)
  }
}

func TestValidateType_InjectionAttempt_ReturnsError(t *testing.T) {
  err := validateType("file OR is_deleted = true")
  if err == nil {
    t.Fatal("expected error for injection attempt in type param")
  }
}

func TestValidateType_SQLInjection_ReturnsError(t *testing.T) {
  err := validateType("' OR 1=1 --")
  if err == nil {
    t.Fatal("expected error for SQL injection attempt in type param")
  }
}

func TestValidateType_ArbitraryString_ReturnsError(t *testing.T) {
  err := validateType("invalid")
  if err == nil {
    t.Fatal("expected error for arbitrary string in type param")
  }
}

// --- validateCheckpointTemplate ---

func TestValidateCheckpointTemplate_PathTraversal_RejectsRelative(t *testing.T) {
  err := validateCheckpointTemplate("../../../etc/passwd")
  if err == nil {
    t.Fatal("expected error for path traversal: ../../../etc/passwd")
  }
}

func TestValidateCheckpointTemplate_AbsolutePath_Rejects(t *testing.T) {
  err := validateCheckpointTemplate("/etc/shadow")
  if err == nil {
    t.Fatal("expected error for absolute path: /etc/shadow")
  }
}

func TestValidateCheckpointTemplate_TraversalInCheckpoints_Rejects(t *testing.T) {
  err := validateCheckpointTemplate("data/checkpoints/../../secrets")
  if err == nil {
    t.Fatal("expected error for traversal within path: data/checkpoints/../../secrets")
  }
}

func TestValidateCheckpointTemplate_ValidPath_Accepts(t *testing.T) {
  err := validateCheckpointTemplate("data/checkpoints/my-checkpoint.json")
  if err != nil {
    t.Fatalf("expected no error for valid path, got: %v", err)
  }
}

func TestValidateCheckpointTemplate_EmptyString_Accepts(t *testing.T) {
  err := validateCheckpointTemplate("")
  if err != nil {
    t.Fatalf("expected no error for empty checkpoint template, got: %v", err)
  }
}
