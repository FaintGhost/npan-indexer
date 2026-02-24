package httpx

import (
	"context"

	"connectrpc.com/connect"

	npanv1 "npan/gen/go/npan/v1"
)

type healthConnectServer struct {
	handlers *Handlers
}

func newHealthConnectServer(handlers *Handlers) *healthConnectServer {
	return &healthConnectServer{handlers: handlers}
}

func (s *healthConnectServer) Health(_ context.Context, _ *connect.Request[npanv1.HealthRequest]) (*connect.Response[npanv1.HealthResponse], error) {
	running := false
	if s.handlers != nil && s.handlers.syncManager != nil {
		running = s.handlers.syncManager.IsRunning()
	}
	return connect.NewResponse(&npanv1.HealthResponse{
		Status:      "ok",
		RunningSync: running,
	}), nil
}

func (s *healthConnectServer) Readyz(_ context.Context, _ *connect.Request[npanv1.ReadyzRequest]) (*connect.Response[npanv1.ReadyzResponse], error) {
	if s.handlers != nil && s.handlers.queryService != nil {
		if err := s.handlers.queryService.Ping(); err != nil {
			meili := "unreachable"
			return connect.NewResponse(&npanv1.ReadyzResponse{
				Status: npanv1.ReadyStatus_READY_STATUS_NOT_READY,
				Meili:  &meili,
			}), nil
		}
	}
	return connect.NewResponse(&npanv1.ReadyzResponse{
		Status: npanv1.ReadyStatus_READY_STATUS_READY,
	}), nil
}
