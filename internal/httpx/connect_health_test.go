package httpx

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	npanv1 "npan/gen/go/npan/v1"
	"npan/gen/go/npan/v1/npanv1connect"
)

func TestConnectHealth_ReturnsRunningSyncState(t *testing.T) {
	t.Parallel()

	handlers := newTestHandlers(t)
	e := NewServer(handlers, testAdminKey, testDistFS(), nil)
	ts := httptest.NewServer(e)
	defer ts.Close()

	client := npanv1connect.NewHealthServiceClient(ts.Client(), ts.URL)
	resp, err := client.Health(context.Background(), connect.NewRequest(&npanv1.HealthRequest{}))
	if err != nil {
		t.Fatalf("Health RPC returned error: %v", err)
	}
	if resp.Msg.GetStatus() != "ok" {
		t.Fatalf("expected status=ok, got %q", resp.Msg.GetStatus())
	}
}

func TestConnectReadyz_ReturnsReadyAndNotReady(t *testing.T) {
	t.Parallel()

	t.Run("ready", func(t *testing.T) {
		handlers := newTestHandlers(t)
		e := NewServer(handlers, testAdminKey, testDistFS(), nil)
		ts := httptest.NewServer(e)
		defer ts.Close()

		client := npanv1connect.NewHealthServiceClient(ts.Client(), ts.URL)
		resp, err := client.Readyz(context.Background(), connect.NewRequest(&npanv1.ReadyzRequest{}))
		if err != nil {
			t.Fatalf("Readyz RPC returned error: %v", err)
		}
		if got := resp.Msg.GetStatus(); got != npanv1.ReadyStatus_READY_STATUS_READY {
			t.Fatalf("expected ready status, got %v", got)
		}
	})

	t.Run("not_ready", func(t *testing.T) {
		handlers := newTestHandlers(t)
		handlers.queryService = &mockSearchService{pingErr: errors.New("meili down")}
		e := NewServer(handlers, testAdminKey, testDistFS(), nil)
		ts := httptest.NewServer(e)
		defer ts.Close()

		client := npanv1connect.NewHealthServiceClient(ts.Client(), ts.URL)
		resp, err := client.Readyz(context.Background(), connect.NewRequest(&npanv1.ReadyzRequest{}))
		if err != nil {
			t.Fatalf("Readyz RPC returned error: %v", err)
		}
		if got := resp.Msg.GetStatus(); got != npanv1.ReadyStatus_READY_STATUS_NOT_READY {
			t.Fatalf("expected not_ready status, got %v", got)
		}
		if got := resp.Msg.GetMeili(); got != "unreachable" {
			t.Fatalf("expected meili=unreachable, got %q", got)
		}
	})
}
