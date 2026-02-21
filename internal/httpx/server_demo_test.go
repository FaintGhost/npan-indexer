package httpx

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestNewServer_RegistersAppRoutes(t *testing.T) {
	t.Parallel()

	e := NewServer(&Handlers{}, "test-key")
	routes := e.Router().Routes()
	seen := map[string]bool{}
	for _, route := range routes {
		seen[route.Method+" "+route.Path] = true
	}

	if !seen["GET /app"] {
		t.Fatal("expected GET /app route to be registered")
	}
	if !seen["GET /app/"] {
		t.Fatal("expected GET /app/ route to be registered")
	}
	if !seen["GET /api/v1/app/search"] {
		t.Fatal("expected GET /api/v1/app/search route to be registered")
	}
	if !seen["GET /api/v1/app/download-url"] {
		t.Fatal("expected GET /api/v1/app/download-url route to be registered")
	}
}

func TestAppPage_ReturnsHTML(t *testing.T) {
	t.Parallel()

	path := resolveAppHTMLPath()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("app html path not found: %s, err=%v", path, err)
	}

	e := NewServer(&Handlers{}, "test-key")
	req := httptest.NewRequest(http.MethodGet, "/app", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(strings.ToLower(body), "<!doctype html>") {
		t.Fatalf("expected html doctype, got: %q", body)
	}
}
