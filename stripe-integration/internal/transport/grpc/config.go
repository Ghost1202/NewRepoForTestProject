package grpc

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	RecoverEnabled bool   `envconfig:"GRPC_RECOVER_ENABLED" default:"true"`
	RecoverStack   bool   `envconfig:"GRPC_RECOVER_STACK" default:"true"`
	RecoverMessage string `envconfig:"GRPC_RECOVER_MESSAGE" default:"internal server error"`
	TracingEnabled bool   `envconfig:"GRPC_TRACING_ENABLED" default:"true"`

	Host             string        `envconfig:"GRPC_HOST" default:"0.0.0.0"`
	Port             int           `envconfig:"GRPC_PORT" default:"9090"`
	Reflection       bool          `envconfig:"GRPC_REFLECTION" default:"true"`
	MaxRecvMsgSize   int           `envconfig:"GRPC_MAX_RECV_MSG_SIZE" default:"4194304"`
	MaxSendMsgSize   int           `envconfig:"GRPC_MAX_SEND_MSG_SIZE" default:"4194304"`
	KeepaliveTime    time.Duration `envconfig:"GRPC_KEEPALIVE_TIME" default:"30s"`
	KeepaliveTimeout time.Duration `envconfig:"GRPC_KEEPALIVE_TIMEOUT" default:"10s"`
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.RecoverMessage, validation.Required),
		validation.Field(&cfg.Host, validation.Required),
		validation.Field(&cfg.Port, validation.Required),
		validation.Field(&cfg.MaxRecvMsgSize, validation.Required),
		validation.Field(&cfg.MaxSendMsgSize, validation.Required),
		validation.Field(&cfg.KeepaliveTime, validation.Required),
		validation.Field(&cfg.KeepaliveTimeout, validation.Required),
	)
}
