package middleware

import (
	"context"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/turtlepavlo/event-searching/internal/domain"
	"github.com/turtlepavlo/event-searching/internal/lib/telemetry"
)

type UserContextKey struct{}

const (
	authorization = "Authorization"
	bearer        = "Bearer"
)

type TokenValidator interface {
	Validate(ctx context.Context, token string) (*domain.User, error)
}

func Auth(validator TokenValidator, log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			authHeader := r.Header.Get(authorization)
			if authHeader == "" {
				telemetry.WithTrace(ctx, log).Warn("auth header missing",
					zap.String("path", r.URL.Path),
				)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != bearer {
				telemetry.WithTrace(ctx, log).Warn("invalid auth header format",
					zap.String("path", r.URL.Path),
				)
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
				attribute.String("user_role", user.Role),
			)

			ctx = context.WithValue(ctx, UserContextKey{}, *user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromContext(ctx context.Context) (domain.User, bool) {
	u, ok := ctx.Value(UserContextKey{}).(domain.User)
	return u, ok
}
