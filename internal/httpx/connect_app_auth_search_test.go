package httpx

import (
	"context"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/reflect/protoreflect"

	npanv1 "npan/gen/go/npan/v1"
	"npan/gen/go/npan/v1/npanv1connect"
)

func requireAppServiceDescriptor(t *testing.T) protoreflect.ServiceDescriptor {
	t.Helper()

	service := npanv1.File_npan_v1_api_proto.Services().ByName(protoreflect.Name("AppService"))
	if service == nil {
		t.Fatal("expected AppService descriptor to exist")
	}
	return service
}

func requireTopLevelMessageDescriptor(t *testing.T, name string) protoreflect.MessageDescriptor {
	t.Helper()

	message := npanv1.File_npan_v1_api_proto.Messages().ByName(protoreflect.Name(name))
	if message == nil {
		t.Fatalf("expected %s message descriptor for public search config contract", name)
	}
	return message
}

func TestAppServiceDescriptor_ExposesGetSearchConfigRPC(t *testing.T) {
	t.Parallel()

	appService := requireAppServiceDescriptor(t)
	method := appService.Methods().ByName(protoreflect.Name("GetSearchConfig"))
	if method == nil {
		t.Fatal("expected AppService to expose GetSearchConfig RPC for public search config bootstrap")
	}

	if got := string(method.Input().Name()); got != "GetSearchConfigRequest" {
		t.Fatalf("expected GetSearchConfig input message, got %q", got)
	}
	if got := string(method.Output().Name()); got != "GetSearchConfigResponse" {
		t.Fatalf("expected GetSearchConfig output message, got %q", got)
	}
}

func TestGetSearchConfigResponseDescriptor_ExposesPublicFieldsOnly(t *testing.T) {
	t.Parallel()

	message := requireTopLevelMessageDescriptor(t, "GetSearchConfigResponse")
	cases := []string{"host", "index_name", "search_api_key", "instantsearch_enabled"}
	for _, fieldName := range cases {
		field := message.Fields().ByName(protoreflect.Name(fieldName))
		if field == nil {
			t.Fatalf("expected GetSearchConfigResponse.%s field for public search config contract", fieldName)
		}
	}
}

func TestGetSearchConfigResponseDescriptor_DoesNotExposePrivateMeiliCredentials(t *testing.T) {
	t.Parallel()

	message := requireTopLevelMessageDescriptor(t, "GetSearchConfigResponse")
	for _, forbidden := range []string{"meili_api_key", "admin_api_key", "master_key", "private_api_key"} {
		if field := message.Fields().ByName(protoreflect.Name(forbidden)); field != nil {
			t.Fatalf("GetSearchConfigResponse must not expose private credential field %q", forbidden)
		}
	}
}

func TestConnectAppGetSearchConfig_ReturnsDedicatedPublicSearchConfig(t *testing.T) {
	t.Parallel()

	handlers := newTestHandlers(t)
	handlers.cfg.PublicSearchHost = "https://search.example.com"
	handlers.cfg.PublicSearchIndexName = "npan-public"
	handlers.cfg.PublicSearchAPIKey = "public-search-key"
	handlers.cfg.PublicSearchInstantsearchOn = true
	handlers.cfg.MeiliAPIKey = "private-meili-key"

	e := NewServer(handlers, testAdminKey, testDistFS(), nil)
	ts := httptest.NewServer(e)
	defer ts.Close()

	client := npanv1connect.NewAppServiceClient(ts.Client(), ts.URL)
	resp, err := client.GetSearchConfig(context.Background(), connect.NewRequest(&npanv1.GetSearchConfigRequest{}))
	if err != nil {
		t.Fatalf("GetSearchConfig RPC returned error: %v", err)
	}
	if got := resp.Msg.GetHost(); got != "https://search.example.com" {
		t.Fatalf("expected public host, got %q", got)
	}
	if got := resp.Msg.GetIndexName(); got != "npan-public" {
		t.Fatalf("expected public index name, got %q", got)
	}
	if got := resp.Msg.GetSearchApiKey(); got != "public-search-key" {
		t.Fatalf("expected dedicated public search key, got %q", got)
	}
	if !resp.Msg.GetInstantsearchEnabled() {
		t.Fatal("expected instantsearch_enabled=true when dedicated public config is complete")
	}
	if resp.Msg.GetSearchApiKey() == handlers.cfg.MeiliAPIKey {
		t.Fatal("expected GetSearchConfig to avoid exposing private MEILI_API_KEY")
	}
}

func TestConnectAppGetSearchConfig_DisablesInstantsearchWhenPublicConfigIncomplete(t *testing.T) {
	t.Parallel()

	handlers := newTestHandlers(t)
	handlers.cfg.PublicSearchHost = "https://search.example.com"
	handlers.cfg.PublicSearchIndexName = "npan-public"
	handlers.cfg.PublicSearchInstantsearchOn = true
	handlers.cfg.MeiliAPIKey = "private-meili-key"

	e := NewServer(handlers, testAdminKey, testDistFS(), nil)
	ts := httptest.NewServer(e)
	defer ts.Close()

	client := npanv1connect.NewAppServiceClient(ts.Client(), ts.URL)
	resp, err := client.GetSearchConfig(context.Background(), connect.NewRequest(&npanv1.GetSearchConfigRequest{}))
	if err != nil {
		t.Fatalf("GetSearchConfig RPC returned error: %v", err)
	}
	if resp.Msg.GetInstantsearchEnabled() {
		t.Fatal("expected instantsearch_enabled=false when dedicated public search config is incomplete")
	}
	if resp.Msg.GetSearchApiKey() != "" {
		t.Fatalf("expected empty searchApiKey when public config is incomplete, got %q", resp.Msg.GetSearchApiKey())
	}
	if resp.Msg.GetSearchApiKey() == handlers.cfg.MeiliAPIKey {
		t.Fatal("expected GetSearchConfig not to fall back to private MEILI_API_KEY")
	}
}

func TestConnectAppGetSearchConfig_DisablesInstantsearchForTypesenseBackend(t *testing.T) {
	t.Parallel()

	handlers := newTestHandlers(t)
	handlers.cfg.SearchBackend = "typesense"
	handlers.cfg.PublicSearchHost = "https://search.example.com"
	handlers.cfg.PublicSearchIndexName = "npan-public"
	handlers.cfg.PublicSearchAPIKey = "public-search-key"
	handlers.cfg.PublicSearchInstantsearchOn = true

	e := NewServer(handlers, testAdminKey, testDistFS(), nil)
	ts := httptest.NewServer(e)
	defer ts.Close()

	client := npanv1connect.NewAppServiceClient(ts.Client(), ts.URL)
	resp, err := client.GetSearchConfig(context.Background(), connect.NewRequest(&npanv1.GetSearchConfigRequest{}))
	if err != nil {
		t.Fatalf("GetSearchConfig RPC returned error: %v", err)
	}
	if resp.Msg.GetInstantsearchEnabled() {
		t.Fatal("expected instantsearch_enabled=false for typesense backend")
	}
}

func TestConnectAppSearch_NoAPIKeyRequired(t *testing.T) {
	t.Parallel()

	handlers := newTestHandlers(t)
	e := NewServer(handlers, testAdminKey, testDistFS(), nil)
	ts := httptest.NewServer(e)
	defer ts.Close()

	client := npanv1connect.NewAppServiceClient(ts.Client(), ts.URL)
	resp, err := client.AppSearch(context.Background(), connect.NewRequest(&npanv1.AppSearchRequest{
		Query: "demo",
	}))
	if err != nil {
		t.Fatalf("AppSearch RPC returned error: %v", err)
	}
	if resp.Msg.GetResult() == nil {
		t.Fatalf("expected result payload")
	}
}

func TestConnectSearchLocal_RequiresAPIKey(t *testing.T) {
	t.Parallel()

	handlers := newTestHandlers(t)
	e := NewServer(handlers, testAdminKey, testDistFS(), nil)
	ts := httptest.NewServer(e)
	defer ts.Close()
	client := npanv1connect.NewSearchServiceClient(ts.Client(), ts.URL)

	_, err := client.LocalSearch(context.Background(), connect.NewRequest(&npanv1.LocalSearchRequest{
		Query: "demo",
	}))
	if err == nil {
		t.Fatalf("expected LocalSearch without API key to fail")
	}

	req := connect.NewRequest(&npanv1.LocalSearchRequest{Query: "demo"})
	req.Header().Set("X-API-Key", testAdminKey)
	resp, err := client.LocalSearch(context.Background(), req)
	if err != nil {
		t.Fatalf("LocalSearch RPC with API key returned error: %v", err)
	}
	if resp.Msg.GetResult() == nil {
		t.Fatalf("expected result payload")
	}
}

func TestConnectAuthCreateToken_ValidatesPayload(t *testing.T) {
	t.Parallel()

	handlers := newTestHandlers(t)
	e := NewServer(handlers, testAdminKey, testDistFS(), nil)
	ts := httptest.NewServer(e)
	defer ts.Close()
	client := npanv1connect.NewAuthServiceClient(ts.Client(), ts.URL)

	req := connect.NewRequest(&npanv1.CreateTokenRequest{})
	req.Header().Set("X-API-Key", testAdminKey)
	_, err := client.CreateToken(context.Background(), req)
	if err == nil {
		t.Fatalf("expected CreateToken with empty payload to fail")
	}
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("expected invalid_argument, got %v", connect.CodeOf(err))
	}
}
