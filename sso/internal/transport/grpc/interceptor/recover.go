package interceptor

import (
	"context"
	"runtime/debug"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Recover(cfg Config, logger *zap.Logger) grpc.UnaryServerInterceptor {
	if !cfg.RecoverEnabled {
		return func(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return handler(ctx, request)
		}
	}

	if logger == nil || logger.Core() == nil {
		logger = zap.NewNop()
	}

	baseLogger := logger.With(
		zap.String("layer", "transport"),
		zap.String("component", "grpc_recover"),
	)

	return func(
		ctx context.Context,
		request any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (response any, handlerError error) {
		defer func() {
			panicValue := recover()
			if panicValue == nil {
				return
			}

			logFields := []zap.Field{
				zap.String("grpc.method", info.FullMethod),
				zap.Any("panic", panicValue),
			}

			if cfg.RecoverStack {
				logFields = append(logFields, zap.ByteString("stack", debug.Stack()))
			}

			baseLogger.Error("panic recovered in grpc handler", logFields...)
			handlerError = status.Error(codes.Internal, cfg.RecoverMessage)
			response = nil
		}()

		return handler(ctx, request)
	}
}
