package interceptor

import (
	"go.uber.org/zap"
	"google.golang.org/grpc"

	config "github.com/turtlepavlo/stripe_integration/internal/transport/grpc"
)

func UnaryServerOption(cfg config.Config, jwtSecret []byte, logger *zap.Logger) grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(
		Logging(logger),
		Recover(cfg, logger),
	)
}
