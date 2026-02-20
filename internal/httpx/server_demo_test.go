package httpx

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestNewServer_RegistersDemoRoutes(t *testing.T) {
	t.Parallel()

	e := NewServer(&Handlers{})
	routes := e.Router().Routes()
	seen := map[string]bool{}
	for _, route := range routes {
		seen[route.Method+" "+route.Path] = true
	}

	if !seen["GET /demo"] {
		t.Fatal("expected GET /demo route to be registered")
	}
	if !seen["GET /demo/"] {
		t.Fatal("expected GET /demo/ route to be registered")
	}
}

func TestDemoPage_ReturnsHTML(t *testing.T) {
	t.Parallel()

	path := resolveDemoHTMLPath()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("demo html path not found: %s, err=%v", path, err)
	}

	e := NewServer(&Handlers{})
	req := httptest.NewRequest(http.MethodGet, "/demo", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(strings.ToLower(body), "<!doctype html>") {
		t.Fatalf("expected html doctype, got: %q", body)
	}
	if !strings.Contains(body, "Npan Search Demo") {
		t.Fatalf("expected demo page title marker, got: %q", body)
	}
}
