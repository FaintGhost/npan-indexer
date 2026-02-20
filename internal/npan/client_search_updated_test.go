package npan

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchUpdatedWindow_DefaultQueryWords(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/item/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("query_words") != "* OR *" {
			t.Fatalf("expected default query_words '* OR *', got %q", query.Get("query_words"))
		}
		if query.Get("type") != "all" {
			t.Fatalf("expected type=all, got %q", query.Get("type"))
		}
		if query.Get("query_filter") != "all" {
			t.Fatalf("expected query_filter=all, got %q", query.Get("query_filter"))
		}
		if query.Get("updated_time_range") != "10,20" {
			t.Fatalf("expected updated_time_range 10,20, got %q", query.Get("updated_time_range"))
		}
		if query.Get("page_id") != "3" {
			t.Fatalf("expected page_id=3, got %q", query.Get("page_id"))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"files":[],"folders":[],"page_count":1}`))
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientOptions{
		BaseURL: server.URL,
		Token:   "test-token",
	})

	start := int64(10)
	end := int64(20)
	if _, err := client.SearchUpdatedWindow(context.Background(), "", &start, &end, 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSearchUpdatedWindow_CustomQueryWords(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("query_words") != "MX40" {
			t.Fatalf("expected query_words MX40, got %q", query.Get("query_words"))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"files":[],"folders":[],"page_count":1}`))
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientOptions{
		BaseURL: server.URL,
		Token:   "test-token",
	})

	if _, err := client.SearchUpdatedWindow(context.Background(), "MX40", nil, nil, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
