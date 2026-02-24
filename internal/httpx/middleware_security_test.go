package httpx

import (
	"testing"
)

func TestCORSConfig_ConnectHeadersIncluded(t *testing.T) {
	t.Parallel()

	cfg := CORSConfig([]string{"https://admin.example.com"})

	assertContainsHeader(t, cfg.AllowHeaders, "Authorization")
	assertContainsHeader(t, cfg.AllowHeaders, "X-API-Key")
	assertContainsHeader(t, cfg.AllowHeaders, "Content-Type")
	assertContainsHeader(t, cfg.AllowHeaders, "Connect-Protocol-Version")
	assertContainsHeader(t, cfg.AllowHeaders, "Connect-Timeout-Ms")
	assertContainsHeader(t, cfg.AllowHeaders, "Grpc-Timeout")
	assertContainsHeader(t, cfg.ExposeHeaders, "Connect-Error-Reason")
	assertContainsHeader(t, cfg.ExposeHeaders, "Connect-Error-Details")
}

func TestParseCORSOrigins_TrimEmpty(t *testing.T) {
	t.Parallel()

	got := ParseCORSOrigins(" https://a.example.com, ,https://b.example.com  ,,")
	if len(got) != 2 {
		t.Fatalf("expected 2 origins, got %d (%v)", len(got), got)
	}
	if got[0] != "https://a.example.com" || got[1] != "https://b.example.com" {
		t.Fatalf("unexpected origins: %v", got)
	}
}

func assertContainsHeader(t *testing.T, headers []string, want string) {
	t.Helper()
	for _, h := range headers {
		if h == want {
			return
		}
	}
	t.Fatalf("header %q not found in %v", want, headers)
}
