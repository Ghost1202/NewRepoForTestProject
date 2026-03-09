package interceptor

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func Logging(logger *zap.Logger) grpc.UnaryServerInterceptor {
	if logger == nil {
		logger = zap.NewNop()
	}

	log := logger.With(
		zap.String("layer", "transport"),
		zap.String("component", "grpc_logging"),
	)

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		fields := []zap.Field{
			zap.String("grpc.method", info.FullMethod),
			zap.String("grpc.code", status.Code(err).String()),
			zap.Duration("duration", duration),
		}

		if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
			fields = append(fields, zap.String("peer.addr", p.Addr.String()))
		}

		if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
			fields = append(fields,
				zap.String("trace_id", sc.TraceID().String()),
				zap.String("span_id", sc.SpanID().String()),
			)
		}

		if err != nil {
			fields = append(fields, zap.Error(err))
			log.Warn("grpc request failed", fields...)
		} else {
			log.Info("grpc request finished", fields...)
		}

		return resp, err
	}
}
