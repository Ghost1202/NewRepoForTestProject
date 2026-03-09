package middleware

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/turtlepavlo/event-service/internal/lib/telemetry"
	transport "github.com/turtlepavlo/event-service/internal/transport/http"
)

func AccessLog(cfg transport.Config, zapLog *zap.Logger) func(http.Handler) http.Handler {
	tracer := otel.Tracer("event-service/internal/transport/http/middleware/logger")

	if !cfg.LoggerEnabled {
		return func(next http.Handler) http.Handler { return next }
	}
	if zapLog == nil {
		zapLog = zap.NewNop()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			ctx, span := tracer.Start(request.Context(), "HTTP Request",
				trace.WithAttributes(
					attribute.String("http.method", request.Method),
					attribute.String("http.url", request.URL.Path),
				),
			)
			defer span.End()

			request = request.WithContext(ctx)

			startTime := time.Now()
			next.ServeHTTP(responseWriter, request)

			telemetry.WithTrace(ctx, zapLog.Named("http")).Info("request processed",
				zap.String("method", request.Method),
				zap.String("path", request.URL.Path),
				zap.String("query", request.URL.RawQuery),
				zap.Duration("duration", time.Since(startTime)),
				zap.String("user_agent", request.UserAgent()),
			)
		})
	}
}
