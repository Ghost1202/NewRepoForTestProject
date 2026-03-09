package interceptor

import (
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func UnaryServerOption(cfg Config, jwtSecret []byte, logger *zap.Logger) grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(
		Logging(logger),
		Recover(cfg, logger),
		RateLimit(cfg, logger),
		TokenValidation(cfg, jwtSecret, logger),
	)
}
