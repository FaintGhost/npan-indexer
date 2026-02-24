package httpx

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	npanv1 "npan/gen/go/npan/v1"
)

func TestConnectValidationInterceptor_AdminStartSyncHitRule(t *testing.T) {
	t.Parallel()

	interceptor := NewConnectValidationInterceptor(nil)
	nextCalled := false
	unary := interceptor.WrapUnary(func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		nextCalled = true
		return connect.NewResponse(&npanv1.StartSyncResponse{Message: "ok"}), nil
	})

	rootWorkers := int64(0)
	_, err := unary(context.Background(), connect.NewRequest(&npanv1.StartSyncRequest{
		RootWorkers: &rootWorkers,
	}))
	if err == nil {
		t.Fatalf("expected invalid_argument error")
	}
	if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
		t.Fatalf("expected invalid_argument, got %v", got)
	}
	if nextCalled {
		t.Fatalf("expected next handler not called when validation fails")
	}
}

func TestConnectValidationInterceptor_PaginationHitRule(t *testing.T) {
	t.Parallel()

	interceptor := NewConnectValidationInterceptor(nil)
	nextCalled := false
	unary := interceptor.WrapUnary(func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		nextCalled = true
		return connect.NewResponse(&npanv1.AppSearchResponse{}), nil
	})

	pageSize := int64(0)
	_, err := unary(context.Background(), connect.NewRequest(&npanv1.AppSearchRequest{
		Query:    "demo",
		PageSize: &pageSize,
	}))
	if err == nil {
		t.Fatalf("expected invalid_argument error")
	}
	if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
		t.Fatalf("expected invalid_argument, got %v", got)
	}
	if nextCalled {
		t.Fatalf("expected next handler not called when validation fails")
	}
}

func TestConnectValidationInterceptor_NoRuleMessageNoop(t *testing.T) {
	t.Parallel()

	interceptor := NewConnectValidationInterceptor(nil)
	nextCalled := false
	unary := interceptor.WrapUnary(func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		nextCalled = true
		return connect.NewResponse(&npanv1.HealthResponse{Status: "ok"}), nil
	})

	_, err := unary(context.Background(), connect.NewRequest(&npanv1.HealthRequest{}))
	if err != nil {
		t.Fatalf("expected no error for message without validation rules, got %v", err)
	}
	if !nextCalled {
		t.Fatalf("expected next handler called for message without validation rules")
	}
}
