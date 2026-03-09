package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	config "github.com/turtlepavlo/event-booking/internal/transport/http"
	"github.com/turtlepavlo/event-booking/pkg/telemetry"
)

func Recover(config config.Config, zapLog *zap.Logger) func(http.Handler) http.Handler {
	if !config.RecoverEnabled {
		return func(next http.Handler) http.Handler { return next }
	}

	if zapLog == nil {
		zapLog = zap.NewNop()
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
					if config.RecoverStack {
						fields = append(fields, zap.ByteString("stack", debug.Stack()))
					}

					logger.Error("panic in handler", fields...)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)

					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(http.StatusInternalServerError)

					if _, writeErr := fmt.Fprintf(w, `{"error":"%s"}`, config.RecoverMessage); writeErr != nil {
						return
					}
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
