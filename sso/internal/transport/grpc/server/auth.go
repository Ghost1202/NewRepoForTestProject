package server

import (
	"context"
	"strings"

	authv1 "github.com/turtlepavlo/proto-contract/gen/go/sso/auth/v1"
	jwtlib "github.com/turtlepavlo/sso/internal/lib/jwt"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) ValidateToken(ctx context.Context, request *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	rawToken := strings.TrimSpace(request.GetAccessToken())
	if rawToken == "" {
		return nil, status.Error(codes.InvalidArgument, "access_token is required")
	}

	tokenString := stripBearerPrefix(rawToken)

	claims, parseErr := jwtlib.ParseToken(tokenString, server.jwtSecret)
	if parseErr != nil {
		server.log.Info("token invalid",
			zap.String("reason", "parse_failed"),
			zap.Error(parseErr),
		)
		return &authv1.ValidateTokenResponse{Valid: false, User: nil}, nil
	}

	domainUser, claimsErr := jwtlib.ParseUserClaims(claims)
	if claimsErr != nil {
		server.log.Info("token invalid",
			zap.String("reason", "claims_invalid"),
			zap.Error(claimsErr),
		)
		return &authv1.ValidateTokenResponse{Valid: false, User: nil}, nil
	}

	return &authv1.ValidateTokenResponse{
		Valid: true,
		User:  toProtoUserFromDomain(domainUser),
	}, nil
}

func stripBearerPrefix(raw string) string {
	parts := strings.Fields(raw)
	if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") && parts[1] != "" {
		return parts[1]
	}
	return raw
}
