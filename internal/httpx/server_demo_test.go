package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewServer_RegistersAppRoutes(t *testing.T) {
	t.Parallel()

	e := NewServer(&Handlers{}, "test-key", testDistFS(), nil)
	routes := e.Router().Routes()
	seen := map[string]bool{}
	for _, route := range routes {
		seen[route.Method+" "+route.Path] = true
	}

	if !seen["GET /*"] {
		t.Fatal("expected GET /* route to be registered")
	}
	if !seen["echo_route_any /npan.v1.AppService/*"] {
		t.Fatalf("expected AppService connect route group to be registered, got routes: %#v", seen)
	}
	if !seen["echo_route_any /npan.v1.SearchService/*"] {
		t.Fatal("expected SearchService connect route group to be registered")
	}
	if !seen["echo_route_any /npan.v1.AuthService/*"] {
		t.Fatal("expected AuthService connect route group to be registered")
	}
}

func TestAppPage_ReturnsHTML(t *testing.T) {
	t.Parallel()

	e := NewServer(&Handlers{}, "test-key", testDistFS(), nil)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
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
