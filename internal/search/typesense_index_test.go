package search

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"npan/internal/models"
)

func TestTypesenseEnsureSettingsCreatesCollectionWhenMissing(t *testing.T) {
	t.Parallel()

	var (
		mu         sync.Mutex
		created    bool
		schemaBody string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/collections/npan_items":
			http.NotFound(w, r)
		case r.Method == http.MethodPost && r.URL.Path == "/collections":
			body, _ := io.ReadAll(r.Body)
			mu.Lock()
			created = true
			schemaBody = string(body)
			mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"name":"npan_items","num_documents":0,"fields":[{"name":"doc_id","type":"string"}]}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	idx := NewTypesenseIndex(srv.URL, "typesense-key", "npan_items")
	if err := idx.EnsureSettings(context.Background()); err != nil {
		t.Fatalf("EnsureSettings returned error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if !created {
		t.Fatal("expected collection creation request")
	}
	for _, required := range []string{`"name":"doc_id"`, `"name":"name_base"`, `"name":"modified_at"`} {
		if !strings.Contains(schemaBody, required) {
			t.Fatalf("expected schema body to contain %s, got %s", required, schemaBody)
		}
	}
}

func TestTypesenseUpsertDocumentsUsesImportUpsert(t *testing.T) {
	t.Parallel()

	var gotQuery string
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/collections/npan_items/documents/import" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		gotQuery = r.URL.RawQuery
		gotBody = string(body)
		_, _ = w.Write([]byte("{\"success\":true}\n"))
	}))
	defer srv.Close()

	idx := NewTypesenseIndex(srv.URL, "typesense-key", "npan_items")
	err := idx.UpsertDocuments(context.Background(), []models.IndexDocument{{
		DocID:      "file_1",
		SourceID:   1,
		Type:       models.ItemTypeFile,
		Name:       "spec.pdf",
		NameBase:   "spec",
		NameExt:    "pdf",
		PathText:   "/spec.pdf",
		ParentID:   0,
		ModifiedAt: 1700000000,
	}})
	if err != nil {
		t.Fatalf("UpsertDocuments returned error: %v", err)
	}
	if !strings.Contains(gotQuery, "action=upsert") {
		t.Fatalf("expected action=upsert query, got %s", gotQuery)
	}
	if !strings.Contains(gotBody, `"doc_id":"file_1"`) {
		t.Fatalf("expected ndjson payload to contain doc_id, got %s", gotBody)
	}
}

func TestTypesenseSearchBuildsFiltersAndParsesHighlights(t *testing.T) {
	t.Parallel()

	var requests []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/collections/npan_items/documents/search" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		requests = append(requests, r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		if len(requests) == 1 {
			_, _ = w.Write([]byte(`{"found":0,"hits":[]}`))
			return
		}
		_, _ = w.Write([]byte(`{
			"found": 1,
			"hits": [
				{
					"document": {
						"doc_id": "file_9",
						"source_id": 9,
						"type": "file",
						"name": "specifications.pdf",
						"name_base": "specifications",
						"name_ext": "pdf",
						"file_category": "doc",
						"path_text": "/specifications.pdf",
						"parent_id": 2,
						"modified_at": 1700000000,
						"created_at": 1700000000,
						"size": 10,
						"sha1": "",
						"in_trash": false,
						"is_deleted": false
					},
					"highlights": [
						{"field":"name","snippet":"<mark>spec</mark>ifications.pdf"}
					]
				}
			]
		}`))
	}))
	defer srv.Close()

	idx := NewTypesenseIndex(srv.URL, "typesense-key", "npan_items")
	parentID := int64(2)
	results, total, err := idx.Search(models.LocalSearchParams{
		Query:    "spec pdf",
		Type:     "file",
		Page:     1,
		PageSize: 5,
		ParentID: &parentID,
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(requests) != 2 {
		t.Fatalf("expected fallback search, got %d requests", len(requests))
	}
	if !strings.Contains(requests[0], "drop_tokens_threshold=0") || !strings.Contains(requests[1], "drop_tokens_threshold=1") {
		t.Fatalf("expected drop_tokens_threshold fallback, got %v", requests)
	}
	decoded, _ := urlQueryUnescape(requests[0])
	for _, expected := range []string{"query_by=name_base,name_ext,name,path_text", "type:=`file`", "parent_id:=2", "in_trash:=false", "is_deleted:=false"} {
		if !strings.Contains(decoded, expected) {
			t.Fatalf("expected query to contain %s, got %s", expected, decoded)
		}
	}
	if total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}
	if len(results) != 1 || results[0].HighlightedName != "<mark>spec</mark>ifications.pdf" {
		t.Fatalf("unexpected search results: %+v", results)
	}
}

func TestTypesenseDocumentCountReadsCollectionStats(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/collections/npan_items" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		payload := map[string]any{
			"name":          "npan_items",
			"num_documents": 42,
			"fields": []map[string]any{
				{"name": "doc_id", "type": "string"},
				{"name": "source_id", "type": "int64"},
				{"name": "type", "type": "string"},
				{"name": "name", "type": "string"},
				{"name": "name_base", "type": "string"},
				{"name": "name_ext", "type": "string"},
				{"name": "file_category", "type": "string"},
				{"name": "path_text", "type": "string"},
				{"name": "parent_id", "type": "int64"},
				{"name": "modified_at", "type": "int64"},
				{"name": "created_at", "type": "int64"},
				{"name": "size", "type": "int64"},
				{"name": "sha1", "type": "string"},
				{"name": "in_trash", "type": "bool"},
				{"name": "is_deleted", "type": "bool"},
			},
		}
		encoded, _ := json.Marshal(payload)
		_, _ = w.Write(encoded)
	}))
	defer srv.Close()

	idx := NewTypesenseIndex(srv.URL, "typesense-key", "npan_items")
	count, err := idx.DocumentCount(context.Background())
	if err != nil {
		t.Fatalf("DocumentCount returned error: %v", err)
	}
	if count != 42 {
		t.Fatalf("expected count=42, got %d", count)
	}
}

func urlQueryUnescape(raw string) (string, error) {
	return url.QueryUnescape(raw)
}
