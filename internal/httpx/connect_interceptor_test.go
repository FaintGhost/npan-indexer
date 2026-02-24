package httpx

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	npanv1 "npan/gen/go/npan/v1"
	"npan/gen/go/npan/v1/npanv1connect"
)

type interceptorHealthStub struct {
	healthErr error
	readyzErr error
}

func (s interceptorHealthStub) Health(_ context.Context, _ *connect.Request[npanv1.HealthRequest]) (*connect.Response[npanv1.HealthResponse], error) {
	if s.healthErr != nil {
		return nil, s.healthErr
	}
	return connect.NewResponse(&npanv1.HealthResponse{Status: "ok"}), nil
}

func (s interceptorHealthStub) Readyz(_ context.Context, _ *connect.Request[npanv1.ReadyzRequest]) (*connect.Response[npanv1.ReadyzResponse], error) {
	if s.readyzErr != nil {
		return nil, s.readyzErr
	}
	return connect.NewResponse(&npanv1.ReadyzResponse{Status: npanv1.ReadyStatus_READY_STATUS_READY}), nil
}

func TestConnectErrorInterceptor_ConvertsPlainErrorToInternal(t *testing.T) {
	t.Parallel()

	path, handler := npanv1connect.NewHealthServiceHandler(
		interceptorHealthStub{healthErr: errors.New("boom")},
		connect.WithInterceptors(NewConnectErrorInterceptor(nil)),
	)

	mux := http.NewServeMux()
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	defer server.Close()

	client := npanv1connect.NewHealthServiceClient(server.Client(), server.URL)
	_, err := client.Health(context.Background(), connect.NewRequest(&npanv1.HealthRequest{}))
	if err == nil {
		t.Fatalf("expected error")
	}
	if got := connect.CodeOf(err); got != connect.CodeInternal {
		t.Fatalf("expected code internal, got %v", got)
	}
}

func TestConnectErrorInterceptor_PreservesConnectError(t *testing.T) {
	t.Parallel()

	path, handler := npanv1connect.NewHealthServiceHandler(
		interceptorHealthStub{readyzErr: connect.NewError(connect.CodeInvalidArgument, errors.New("bad request"))},
		connect.WithInterceptors(NewConnectErrorInterceptor(nil)),
	)

	mux := http.NewServeMux()
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	defer server.Close()

	client := npanv1connect.NewHealthServiceClient(server.Client(), server.URL)
	_, err := client.Readyz(context.Background(), connect.NewRequest(&npanv1.ReadyzRequest{}))
	if err == nil {
		t.Fatalf("expected error")
	}
	if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
		t.Fatalf("expected code invalid_argument, got %v", got)
	}
}
