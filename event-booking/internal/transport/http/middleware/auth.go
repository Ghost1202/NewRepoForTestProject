package middleware

import (
	"context"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/turtlepavlo/event-booking/internal/domain"
	"github.com/turtlepavlo/event-booking/pkg/telemetry"
)

type UserContextKey struct{}

const (
	authorization = "Authorization"
	bearer        = "Bearer"
)

type TokenValidator interface {
	Validate(ctx context.Context, token string) (*domain.User, error)
}

func Auth(validator TokenValidator, zapLog *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			log := telemetry.WithTrace(ctx, zapLog)

			authHeader := r.Header.Get(authorization)
			if authHeader == "" {
				log.Warn("auth header missing", zap.String("path", r.URL.Path))
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != bearer {
				log.Warn("invalid auth header format", zap.String("path", r.URL.Path))
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			user, err := validator.Validate(ctx, parts[1])
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			span := trace.SpanFromContext(ctx)
			span.SetAttributes(
				attribute.Int64("user_id", user.ID),
			)

			ctx = context.WithValue(ctx, UserContextKey{}, user.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
