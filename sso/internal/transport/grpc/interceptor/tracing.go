package interceptor

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func TracingOption(cfg Config) (grpc.ServerOption, bool) {
	if !cfg.TracingEnabled {
		return nil, false
	}

	return grpc.StatsHandler(otelgrpc.NewServerHandler()), true
}
