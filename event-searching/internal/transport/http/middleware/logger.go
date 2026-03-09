package middleware

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/turtlepavlo/event-searching/internal/lib/telemetry"
	config "github.com/turtlepavlo/event-searching/internal/transport/http"
)

func AccessLog(cfg config.Config, zapLog *zap.Logger) func(http.Handler) http.Handler {
	var tracer = otel.Tracer("event-searching/internal/transport/http/middleware-logger")

	if !cfg.LoggerEnabled {
		return func(next http.Handler) http.Handler { return next }
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracer.Start(r.Context(), "HTTP Request",
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.url", r.URL.Path),
				),
			)
			defer span.End()

			r = r.WithContext(ctx)

			startTime := time.Now()
			next.ServeHTTP(w, r)

			log := telemetry.WithTrace(ctx, zapLog.Named("http"))
			log.Info("request processed",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("query", r.URL.RawQuery),
				zap.Duration("duration", time.Since(startTime)),
				zap.String("user_agent", r.UserAgent()),
			)
		})
	}
}
