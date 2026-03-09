package server

import (
	"context"

	authv1 "github.com/turtlepavlo/proto-contract/gen/go/sso/auth/v1"
	"github.com/turtlepavlo/sso/internal/domain"
	"go.uber.org/zap"
)

type AuthService interface {
	CreateUser(ctx context.Context, in domain.CreateUser) (int64, error)
	Login(ctx context.Context, identifier string, password string, isEmail bool) (domain.AuthSession, error)
}

type Server struct {
	authv1.UnimplementedAuthServiceServer

	authService AuthService
	jwtSecret   []byte
	log         *zap.Logger
}

func New(authService AuthService, jwtSecret []byte, log *zap.Logger) *Server {
	return &Server{
		authService: authService,
		jwtSecret:   jwtSecret,
		log: log.With(
			zap.String("layer", "transport"),
			zap.String("component", "grpc_server"),
		),
	}
}
