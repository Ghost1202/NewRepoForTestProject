package service

import (
	"context"

	"go.opentelemetry.io/otel"
)

func (s *AuthService) GetGoogleAuthURL(ctx context.Context) string {
	const op = "AuthService.GetGoogleAuthURL"
	tracer := otel.Tracer("sso/service/auth")

	ctx, span := tracer.Start(ctx, op)
	defer span.End()

	return s.googleProvider.AuthCodeURL(ctx, "state-string")
}
