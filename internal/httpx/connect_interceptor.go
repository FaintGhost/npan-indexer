package httpx

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
)

// NewConnectErrorInterceptor 统一处理 Connect handler 返回的非 Connect 错误。
// 业务层显式返回的 *connect.Error 会原样透传。
func NewConnectErrorInterceptor(logger *slog.Logger) connect.Interceptor {
	if logger == nil {
		logger = slog.Default()
	}

	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			resp, err := next(ctx, req)
			if err == nil {
				return resp, nil
			}

			var connectErr *connect.Error
			if errors.As(err, &connectErr) {
				return nil, err
			}

			logger.Error("connect handler error", "procedure", req.Spec().Procedure, "error", err)
			return nil, connect.NewError(connect.CodeInternal, errors.New("服务器内部错误"))
		}
	})
}
