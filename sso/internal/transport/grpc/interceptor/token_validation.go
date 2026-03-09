package interceptor

import (
	"context"
	"strings"

	jwtlib "github.com/turtlepavlo/sso/internal/lib/jwt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	headerAuthorization = "authorization"
	schemeBearer        = "Bearer "

	errMissingToken  = "missing authorization token"
	errInvalidToken  = "invalid authorization token"
	errInvalidClaims = "invalid token claims"
)

func TokenValidation(config Config, jwtSecret []byte, logger *zap.Logger) grpc.UnaryServerInterceptor {
	if !config.AuthEnabled {
		return func(serverContext context.Context, serverRequest any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return handler(serverContext, serverRequest)
		}
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	protectedMethods := make(map[string]struct{})
	for _, method := range strings.Split(config.AuthProtectedMethods, ",") {
		if method = strings.TrimSpace(method); method != "" {
			protectedMethods[method] = struct{}{}
		}
	}

	isGlobalProtection := len(protectedMethods) == 0

	return func(serverContext context.Context, serverRequest any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		_, isMethodProtected := protectedMethods[info.FullMethod]
		if !isGlobalProtection && !isMethodProtected {
			return handler(serverContext, serverRequest)
		}

		incomingMetadata, _ := metadata.FromIncomingContext(serverContext)
		headerValues := incomingMetadata.Get(headerAuthorization)

		if len(headerValues) == 0 || headerValues[0] == "" {
			return nil, status.Error(codes.Unauthenticated, errMissingToken)
		}

		rawToken := headerValues[0]
		accessToken := strings.TrimSpace(strings.TrimPrefix(rawToken, schemeBearer))

		if accessToken == "" {
			return nil, status.Error(codes.Unauthenticated, errMissingToken)
		}

		claims, err := jwtlib.ParseToken(accessToken, jwtSecret)
		if err != nil {
			logger.Warn(errInvalidToken, zap.String("method", info.FullMethod), zap.Error(err))
			return nil, status.Error(codes.Unauthenticated, errInvalidToken)
		}

		if _, err := jwtlib.ParseUserClaims(claims); err != nil {
			logger.Warn(errInvalidClaims, zap.String("method", info.FullMethod), zap.Error(err))
			return nil, status.Error(codes.Unauthenticated, errInvalidToken)
		}

		return handler(serverContext, serverRequest)
	}
}
