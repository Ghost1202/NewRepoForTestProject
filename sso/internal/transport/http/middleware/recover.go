package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func Recover(cfg Config, zapLog *zap.Logger) func(http.Handler) http.Handler {
	if !cfg.RecoverEnabled {
		return func(next http.Handler) http.Handler { return next }
	}
	if zapLog == nil {
		zapLog = zap.NewNop()
	}

	withStack := cfg.RecoverStack

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					span := trace.SpanFromContext(r.Context())
					span.SetStatus(codes.Error, "panic")

					span.SetAttributes(
						attribute.String("panic.value", fmt.Sprint(recovered)),
						attribute.String("http.method", r.Method),
						attribute.String("http.route", r.URL.Path),
					)

					span.RecordError(fmt.Errorf("panic: %v", recovered))
					span.AddEvent("panic recovered")

					recoverLogger := zapLog.With(
						zap.String("layer", "transport"),
						zap.String("component", "recover"),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
					)

					if withStack {
						recoverLogger.Error("panic in handler",
							zap.Any("panic", recovered),
							zap.ByteString("stack", debug.Stack()),
						)
					} else {
						recoverLogger.Error("panic in handler",
							zap.Any("panic", recovered),
						)
					}

					w.WriteHeader(http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
