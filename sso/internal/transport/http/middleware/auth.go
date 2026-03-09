package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/turtlepavlo/sso/internal/lib/jwt"
)

const userHeaderKey = "X-User"

func Auth(secret []byte, logger *zap.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = zap.NewNop()
	}

	log := logger.With(
		zap.String("layer", "transport"),
		zap.String("component", "auth_mw"),
		zap.String("op", "middleware.Auth"),
	)

	tr := otel.Tracer("middleware/auth")

	const authorizationHeaderKey = "Authorization"
	const bearerScheme = "Bearer"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Del(userHeaderKey)

			ctx, span := tr.Start(r.Context(), "AuthMW.ValidateJWT")
			defer span.End()
			r = r.WithContext(ctx)

			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", r.URL.Path),
			)

			traceFields := []zap.Field{}
			if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
				traceFields = append(traceFields,
					zap.String("trace_id", sc.TraceID().String()),
					zap.String("span_id", sc.SpanID().String()),
				)
			}

			authHeaderValue := r.Header.Get(authorizationHeaderKey)
			if authHeaderValue == "" {
				span.SetStatus(codes.Error, "missing authorization header")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			parts := strings.Fields(authHeaderValue)
			if len(parts) != 2 || !strings.EqualFold(parts[0], bearerScheme) || parts[1] == "" {
				span.SetStatus(codes.Error, "invalid authorization header format")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			claims, parseErr := jwt.ParseToken(parts[1], secret)
			if parseErr != nil {
				span.RecordError(parseErr)
				span.SetStatus(codes.Error, "invalid token")

				log.Info("unauthorized",
					append(traceFields,
						zap.String("reason", "invalid_token"),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
					)...,
				)

				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			userData, claimsErr := jwt.ParseUserClaims(claims)
			if claimsErr != nil {
				span.RecordError(claimsErr)
				span.SetStatus(codes.Error, "invalid claims")

				log.Info("unauthorized",
					append(traceFields,
						zap.String("reason", "invalid_claims"),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
					)...,
				)

				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			rawUserData, marshalErr := json.Marshal(userData)
			if marshalErr != nil {
				span.RecordError(marshalErr)
				span.SetStatus(codes.Error, "marshal user failed")

				log.Error("internal error",
					append(traceFields,
						zap.String("reason", "marshal_user_failed"),
						zap.Error(marshalErr),
					)...,
				)

				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Header.Set(userHeaderKey, string(rawUserData))

			span.SetStatus(codes.Ok, "ok")
			next.ServeHTTP(w, r)
		})
	}
}
