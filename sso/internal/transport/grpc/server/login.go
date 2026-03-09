package server

import (
	"context"
	"errors"
	"strings"

	authv1 "github.com/turtlepavlo/proto-contract/gen/go/sso/auth/v1"
	"github.com/turtlepavlo/sso/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) Login(ctx context.Context, request *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	identifier := strings.TrimSpace(request.GetIdentifier())
	password := strings.TrimSpace(request.GetPassword())
	isEmail := request.GetIsEmail()

	if identifier == "" || password == "" {
		return nil, status.Error(codes.InvalidArgument, "identifier and password are required")
	}

	result, err := server.authService.Login(ctx, identifier, password, isEmail)
	if err != nil {
		grpcErr, expected := mapLoginError(err)

		if expected {
			server.log.Info("login rejected",
				zap.Bool("is_email", isEmail),
				zap.Error(grpcErr),
			)
			return nil, grpcErr
		}

		server.log.Error("login failed",
			zap.Bool("is_email", isEmail),
			zap.Error(err),
		)
		return nil, grpcErr
	}

	userMessage := toProtoUserFromLoginResult(result, identifier, isEmail)

	server.log.Info("login successful",
		zap.Int64("user_id", result.UserID),
		zap.String("role", result.Role),
		zap.Bool("is_email", isEmail),
	)

	return &authv1.LoginResponse{
		AccessToken: result.AccessToken,
		User:        userMessage,
	}, nil
}

func mapLoginError(err error) (grpcErr error, expected bool) {
	switch {
	case errors.Is(err, service.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid credentials"), true
	default:
		return status.Error(codes.Internal, "internal error"), false
	}
}
