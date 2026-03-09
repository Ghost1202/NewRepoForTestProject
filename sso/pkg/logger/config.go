package logger

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Mode  string `envconfig:"LOG_MODE"  default:"prod"`
	Level string `envconfig:"LOG_LEVEL" default:"info"`

	Encoding    string `envconfig:"LOG_ENCODING" default:"json"`
	Output      string `envconfig:"LOG_OUTPUT" default:"stdout"`
	ErrorOutput string `envconfig:"LOG_ERROR_OUTPUT" default:"stderr"`

	TimeKey    string `envconfig:"LOG_TIME_KEY" default:"ts"`
	LevelKey   string `envconfig:"LOG_LEVEL_KEY" default:"level"`
	MessageKey string `envconfig:"LOG_MESSAGE_KEY" default:"msg"`
	CallerKey  string `envconfig:"LOG_CALLER_KEY" default:"caller"`

	OTLPEnabled  bool   `envconfig:"LOG_OTLP_ENABLED" default:"false"`
	OTLPEndpoint string `envconfig:"LOG_OTLP_ENDPOINT" default:"localhost:4317"`
	Insecure     bool   `envconfig:"LOG_OTLP_INSECURE" default:"true"`

	OTLPScope string `envconfig:"LOG_OTLP_SCOPE"`
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	if !cfg.OTLPEnabled {
		return validation.ValidateStructWithContext(ctx, &cfg,
			validation.Field(&cfg.Mode, validation.Required),
			validation.Field(&cfg.Level, validation.Required),

			validation.Field(&cfg.Encoding, validation.Required),
			validation.Field(&cfg.Output, validation.Required),
			validation.Field(&cfg.ErrorOutput, validation.Required),

			validation.Field(&cfg.TimeKey, validation.Required),
			validation.Field(&cfg.LevelKey, validation.Required),
			validation.Field(&cfg.MessageKey, validation.Required),
			validation.Field(&cfg.CallerKey, validation.Required),
		)
	}

	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.Mode, validation.Required),
		validation.Field(&cfg.Level, validation.Required),

		validation.Field(&cfg.Encoding, validation.Required),
		validation.Field(&cfg.Output, validation.Required),
		validation.Field(&cfg.ErrorOutput, validation.Required),

		validation.Field(&cfg.TimeKey, validation.Required),
		validation.Field(&cfg.LevelKey, validation.Required),
		validation.Field(&cfg.MessageKey, validation.Required),
		validation.Field(&cfg.CallerKey, validation.Required),

		validation.Field(&cfg.OTLPEndpoint, validation.Required),
		validation.Field(&cfg.OTLPScope, validation.Required),
	)
}
