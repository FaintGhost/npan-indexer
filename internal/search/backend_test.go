package search

import "testing"

func TestParseBackend(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		input   string
		want    Backend
		wantErr bool
	}{
		{name: "default", input: "", want: BackendMeilisearch},
		{name: "meili alias", input: "meili", want: BackendMeilisearch},
		{name: "meilisearch", input: "meilisearch", want: BackendMeilisearch},
		{name: "typesense", input: "typesense", want: BackendTypesense},
		{name: "invalid", input: "elastic", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseBackend(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseBackend returned error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("ParseBackend(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSupportsPublicInstantsearch(t *testing.T) {
	t.Parallel()

	if !SupportsPublicInstantsearch(BackendMeilisearch) {
		t.Fatal("expected meilisearch to support public instantsearch")
	}
	if SupportsPublicInstantsearch(BackendTypesense) {
		t.Fatal("expected typesense to disable public instantsearch bootstrap")
	}
}
