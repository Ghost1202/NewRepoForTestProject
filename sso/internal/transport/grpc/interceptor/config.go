package interceptor

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	RecoverEnabled       bool          `envconfig:"GRPC_RECOVER_ENABLED" default:"true"`
	RecoverStack         bool          `envconfig:"GRPC_RECOVER_STACK" default:"true"`
	RecoverMessage       string        `envconfig:"GRPC_RECOVER_MESSAGE" default:"internal server error"`
	TracingEnabled       bool          `envconfig:"GRPC_TRACING_ENABLED" default:"true"`
	AuthEnabled          bool          `envconfig:"GRPC_AUTH_ENABLED" default:"false"`
	AuthProtectedMethods string        `envconfig:"GRPC_AUTH_PROTECTED_METHODS" default:""`
	RateLimitEnabled     bool          `envconfig:"GRPC_RATELIMIT_ENABLED" default:"false"`
	RateLimitRPS         int64         `envconfig:"GRPC_RATELIMIT_RPS" default:"100"`
	RateLimitBurst       int64         `envconfig:"GRPC_RATELIMIT_BURST" default:"200"`
	RateLimitCleanup     time.Duration `envconfig:"GRPC_RATELIMIT_CLEANUP_INTERVAL" default:"1m"`
	RateLimitTTL         time.Duration `envconfig:"GRPC_RATELIMIT_EXPIRATION" default:"5m"`
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.RecoverMessage, validation.Required),
	)
}
