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

func (server *Server) Register(ctx context.Context, request *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	login := strings.TrimSpace(request.GetLogin())
	email := strings.TrimSpace(strings.ToLower(request.GetEmail()))
	create := ToDomainCreateUser(request)
	userID, err := server.authService.CreateUser(ctx, create)
	if err != nil {
		grpcErr, expected := mapRegisterError(err)

		if expected {
			server.log.Info("register rejected",
				zap.String("email", email),
				zap.String("login", login),
				zap.Error(grpcErr),
			)
			return nil, grpcErr
		}

		server.log.Error("register failed",
			zap.String("email", email),
			zap.String("login", login),
			zap.Error(err),
		)
		return nil, grpcErr
	}

	server.log.Info("user registered",
		zap.Int64("user_id", userID),
		zap.String("email", email),
		zap.String("login", login),
	)

	return &authv1.RegisterResponse{Id: userID}, nil
}

func mapRegisterError(err error) (grpcErr error, expected bool) {
	switch {
	case errors.Is(err, service.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, "user already exists"), true
	default:
		return status.Error(codes.Internal, "internal error"), false
	}
}
