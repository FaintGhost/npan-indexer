package npan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetFolderInfo_DirectPayload(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/folder/123/info" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":123,"name":"PIXELHUE","item_count":4151}`))
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientOptions{
		BaseURL: server.URL,
		Token:   "test-token",
	})

	folder, err := client.GetFolderInfo(context.Background(), 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if folder.ID != 123 {
		t.Fatalf("expected ID=123, got %d", folder.ID)
	}
	if folder.Name != "PIXELHUE" {
		t.Fatalf("expected Name=PIXELHUE, got %q", folder.Name)
	}
	if folder.ItemCount != 4151 {
		t.Fatalf("expected ItemCount=4151, got %d", folder.ItemCount)
	}
}

func TestGetFolderInfo_WrappedPayload(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"folder":{"id":456,"name":"MX40","item_count":99}}`))
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientOptions{
		BaseURL: server.URL,
		Token:   "test-token",
	})

	folder, err := client.GetFolderInfo(context.Background(), 456)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if folder.ID != 456 || folder.Name != "MX40" || folder.ItemCount != 99 {
		t.Fatalf("unexpected folder: %#v", folder)
	}
}
