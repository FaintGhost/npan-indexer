package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"

	"npan/internal/config"
	"npan/internal/npan"
)

func newTestContext(method string, target string) *echo.Context {
	e := echo.New()
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestResolveAuthOptions_NoConfigFallback(t *testing.T) {
	h := &Handlers{cfg: config.Config{
		Token:        "server-token",
		ClientID:     "server-client-id",
		ClientSecret: "server-client-secret",
		SubID:        999,
		SubType:      npan.TokenSubjectEnterprise,
		OAuthHost:    "https://oauth.example.com",
	}}

	c := newTestContext(http.MethodGet, "/api/v1/search/remote")
	// No allow_config_fallback set in context, and config default is false
	opts := h.resolveAuthOptions(c, authPayload{})

	if opts.Token != "" {
		t.Fatalf("expected empty token, got %q", opts.Token)
	}
	if opts.ClientID != "" || opts.ClientSecret != "" {
		t.Fatalf("expected empty client credentials, got %q / %q", opts.ClientID, opts.ClientSecret)
	}
	if opts.SubID != 0 {
		t.Fatalf("expected empty sub id, got %d", opts.SubID)
	}
	if opts.SubType != npan.TokenSubjectUser {
		t.Fatalf("expected default sub type user, got %q", opts.SubType)
	}
	if opts.OAuthHost == "" {
		t.Fatalf("expected default oauth host")
	}
}

func TestResolveAuthOptions_WithConfigFallback(t *testing.T) {
	h := &Handlers{cfg: config.Config{
		Token:        "server-token",
		ClientID:     "server-client-id",
		ClientSecret: "server-client-secret",
		SubID:        999,
		SubType:      npan.TokenSubjectEnterprise,
		OAuthHost:    "https://oauth.example.com",
	}}

	c := newTestContext(http.MethodGet, "/api/v1/app/search")
	// Simulate EmbeddedAuth middleware setting the context value
	c.Set("allow_config_fallback", true)
	opts := h.resolveAuthOptions(c, authPayload{})

	if opts.Token != "server-token" {
		t.Fatalf("expected fallback token, got %q", opts.Token)
	}
	if opts.ClientID != "server-client-id" || opts.ClientSecret != "server-client-secret" {
		t.Fatalf("unexpected client credentials: %q / %q", opts.ClientID, opts.ClientSecret)
	}
	if opts.SubID != 999 {
		t.Fatalf("expected sub id 999, got %d", opts.SubID)
	}
	if opts.SubType != npan.TokenSubjectEnterprise {
		t.Fatalf("expected enterprise sub type, got %q", opts.SubType)
	}
	if opts.OAuthHost != "https://oauth.example.com" {
		t.Fatalf("expected fallback oauth host, got %q", opts.OAuthHost)
	}
}

func TestResolveAuthOptions_GlobalConfigFallback(t *testing.T) {
	h := &Handlers{cfg: config.Config{
		Token:                    "server-token",
		AllowConfigAuthFallback: true,
	}}

	c := newTestContext(http.MethodGet, "/api/v1/search/local")
	// No context value set — should use h.cfg.AllowConfigAuthFallback
	opts := h.resolveAuthOptions(c, authPayload{})

	if opts.Token != "server-token" {
		t.Fatalf("expected fallback token from global config, got %q", opts.Token)
	}
}
