package httpx

import (
	"context"
	"errors"
	"log/slog"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
)

// NewConnectValidationInterceptor 对请求消息执行 protovalidate 校验。
// 当前 proto 尚未添加 validate 规则时，该拦截器会以 no-op 方式运行。
func NewConnectValidationInterceptor(logger *slog.Logger) connect.Interceptor {
	if logger == nil {
		logger = slog.Default()
	}

	validator, err := protovalidate.New()
	if err != nil {
		panic("failed to initialize protovalidate: " + err.Error())
	}

	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			msg, ok := req.Any().(proto.Message)
			if ok {
				if validateErr := validator.Validate(msg); validateErr != nil {
					logger.Warn("connect request validation failed", "procedure", req.Spec().Procedure, "error", validateErr)
					return nil, connect.NewError(connect.CodeInvalidArgument, validateErr)
				}
			}
			return next(ctx, req)
		}
	})
}

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
