package middleware

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	CORSEnabled      bool   `envconfig:"MW_CORS_ENABLED" default:"true"`
	CORSAllowOrigins string `envconfig:"MW_CORS_ALLOW_ORIGINS" default:"*"`
	CORSAllowMethods string `envconfig:"MW_CORS_ALLOW_METHODS" default:"GET,POST,PUT,DELETE,OPTIONS"`
	CORSAllowHeaders string `envconfig:"MW_CORS_ALLOW_HEADERS" default:"Content-Type, Authorization"`
	CORSAllowCredentials bool `envconfig:"MW_CORS_ALLOW_CREDENTIALS" default:"false"`

	LoggerEnabled bool `envconfig:"MW_LOGGER_ENABLED" default:"true"`

	RecoverEnabled bool   `envconfig:"MW_RECOVER_ENABLED" default:"true"`
	RecoverStack   bool   `envconfig:"MW_RECOVER_STACK" default:"true"`
	RecoverMessage string `envconfig:"MW_RECOVER_MESSAGE" default:"internal server error"`

	ThrottleEnabled         bool          `envconfig:"MW_THROTTLE_ENABLED" default:"true"`
	ThrottleRatePerSec      float64       `envconfig:"MW_THROTTLE_RATE" default:"10"`
	ThrottleBurst           int64         `envconfig:"MW_THROTTLE_BURST" default:"20"`
	ThrottleCleanupInterval time.Duration `envconfig:"MW_THROTTLE_CLEANUP_INTERVAL" default:"1m"`
	ThrottleExpiration      time.Duration `envconfig:"MW_THROTTLE_EXPIRATION" default:"5m"`
}

func (cfg Config) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.CORSAllowOrigins, validation.When(cfg.CORSEnabled, validation.Required)),
		validation.Field(&cfg.CORSAllowMethods, validation.When(cfg.CORSEnabled, validation.Required)),
		validation.Field(&cfg.CORSAllowHeaders, validation.When(cfg.CORSEnabled, validation.Required)),

		validation.Field(&cfg.RecoverMessage, validation.When(cfg.RecoverEnabled, validation.Required)),

		validation.Field(&cfg.ThrottleRatePerSec, validation.When(cfg.ThrottleEnabled, validation.Required)),
		validation.Field(&cfg.ThrottleBurst, validation.When(cfg.ThrottleEnabled, validation.Required)),
		validation.Field(&cfg.ThrottleCleanupInterval, validation.When(cfg.ThrottleEnabled, validation.Required)),
		validation.Field(&cfg.ThrottleExpiration, validation.When(cfg.ThrottleEnabled, validation.Required)),
	)
}
