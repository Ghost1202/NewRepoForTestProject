package interceptor

import (
	"context"
	"runtime/debug"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	config "github.com/turtlepavlo/stripe_integration/internal/transport/grpc"
	"github.com/turtlepavlo/stripe_integration/pkg/telemetry"
)

func Recover(cfg config.Config, logger *zap.Logger) grpc.UnaryServerInterceptor {
	if !cfg.RecoverEnabled {
		return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return handler(ctx, req)
		}
	}

	if logger == nil {
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
			if panicValue := recover(); panicValue != nil {
				log := telemetry.WithTrace(ctx, baseLogger)

				fields := []zap.Field{
					zap.String("grpc.method", info.FullMethod),
					zap.Any("panic", panicValue),
				}

				if cfg.RecoverStack {
					fields = append(fields, zap.ByteString("stack", debug.Stack()))
				}

				log.Error("panic recovered in grpc handler", fields...)

				handlerError = status.Error(codes.Internal, cfg.RecoverMessage)
				response = nil
			}
		}()

		return handler(ctx, request)
	}
}
