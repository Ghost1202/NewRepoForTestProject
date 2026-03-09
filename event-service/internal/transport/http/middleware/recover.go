package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/turtlepavlo/event-service/internal/lib/telemetry"
	transport "github.com/turtlepavlo/event-service/internal/transport/http"
)

func Recover(cfg transport.Config, zapLog *zap.Logger) func(http.Handler) http.Handler {
	if !cfg.RecoverEnabled {
		return func(next http.Handler) http.Handler { return next }
	}

	errorMessage := cfg.RecoverMessage
	includeStack := cfg.RecoverStack

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			defer func() {
				if recoveredValue := recover(); recoveredValue != nil {
					ctx := request.Context()
					span := trace.SpanFromContext(ctx)
					span.RecordError(fmt.Errorf("%v", recoveredValue))
					span.SetStatus(codes.Error, "panic recovered")

					logger := zapLog.Named("recover").With(
						zap.String("method", request.Method),
						zap.String("path", request.URL.Path),
					)

					if includeStack {
						telemetry.WithTrace(ctx, logger).Error("panic in handler",
							zap.Any("panic", recoveredValue),
							zap.ByteString("stack", debug.Stack()),
						)
					} else {
						telemetry.WithTrace(ctx, logger).Error("panic in handler",
							zap.Any("panic", recoveredValue),
						)
					}

					responseWriter.Header().Set("Content-Type", "application/json")
					responseWriter.WriteHeader(http.StatusInternalServerError)

					errorBody := `{"error":"` + errorMessage + `"}`
					if _, err := responseWriter.Write([]byte(errorBody)); err != nil {
						telemetry.WithTrace(ctx, logger).Warn("failed to write error body", zap.Error(err))
					}
				}
			}()
			next.ServeHTTP(responseWriter, request)
		})
	}
}
