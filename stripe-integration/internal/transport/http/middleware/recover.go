package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	config "github.com/turtlepavlo/stripe_integration/internal/transport/http"
	"github.com/turtlepavlo/stripe_integration/pkg/telemetry"
)

func Recover(cfg config.Config, zapLog *zap.Logger) func(http.Handler) http.Handler {
	if !cfg.RecoverEnabled {
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
					span.RecordError(fmt.Errorf("panic: %v", rec))
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

					logger.Error("panic recovered in http handler", fields...)

					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(http.StatusInternalServerError)

					fmt.Fprintf(w, `{"error":"%s"}`, cfg.RecoverMessage)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
