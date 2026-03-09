package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/turtlepavlo/event-searching/internal/lib/telemetry"
	config "github.com/turtlepavlo/event-searching/internal/transport/http"
)

func Recover(cfg config.Config, zapLog *zap.Logger) func(http.Handler) http.Handler {
	if !cfg.RecoverEnabled {
		return func(next http.Handler) http.Handler { return next }
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					ctx := r.Context()
					span := trace.SpanFromContext(ctx)
					span.RecordError(fmt.Errorf("%v", rec))
					span.SetStatus(codes.Error, "panic recovered")

					logger := zapLog.Named("recover").With(
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
					)

					logger = telemetry.WithTrace(ctx, logger)

					fields := []zap.Field{
						zap.Any("panic", rec),
					}
					if cfg.RecoverStack {
						fields = append(fields, zap.ByteString("stack", debug.Stack()))
					}

					logger.Error("panic in handler", fields...)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					if _, err := fmt.Fprintf(w, `{"error":"%s"}`, cfg.RecoverMessage); err != nil {
						return
					}
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
