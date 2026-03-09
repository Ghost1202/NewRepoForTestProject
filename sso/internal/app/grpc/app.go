package grpcapp

import (
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	authv1 "github.com/turtlepavlo/proto-contract/gen/go/sso/auth/v1"
	"github.com/turtlepavlo/sso/internal/transport/grpc/interceptor"
	"github.com/turtlepavlo/sso/internal/transport/grpc/server"
)

type App struct {
	logger        *zap.Logger
	gRPCServer    *grpc.Server
	serverAddress string
}

func New(
	logger *zap.Logger,
	authService authv1.AuthServiceServer,
	serverConfig server.Config,
	interceptorConfig interceptor.Config,
	jwtSecretKey []byte,
) *App {
	serverOptions := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(serverConfig.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(serverConfig.MaxSendMsgSize),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    serverConfig.KeepaliveTime,
			Timeout: serverConfig.KeepaliveTimeout,
		}),
		interceptor.UnaryServerOption(interceptorConfig, jwtSecretKey, logger),
	}

	if tracingOption, isEnabled := interceptor.TracingOption(interceptorConfig); isEnabled {
		serverOptions = append(serverOptions, tracingOption)
	}

	gRPCServer := grpc.NewServer(serverOptions...)

	authv1.RegisterAuthServiceServer(gRPCServer, authService)

	if serverConfig.Reflection {
		reflection.Register(gRPCServer)
	}

	return &App{
		logger:        logger,
		gRPCServer:    gRPCServer,
		serverAddress: fmt.Sprintf("%s:%d", serverConfig.Host, serverConfig.Port),
	}
}

func (application *App) Run() error {
	const operation = "grpcapp.Run"

	listener, err := net.Listen("tcp", application.serverAddress)
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}

	application.logger.Info("grpc server is running", zap.String("address", application.serverAddress))

	if err := application.gRPCServer.Serve(listener); err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}

	return nil
}

func (application *App) Stop() {
	const operation = "grpcapp.Stop"

	application.logger.Info("stopping grpc server", zap.String("operation", operation))
	application.gRPCServer.GracefulStop()
}
