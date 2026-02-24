package httpx

import (
	"context"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	npanv1 "npan/gen/go/npan/v1"
	"npan/gen/go/npan/v1/npanv1connect"
)

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
