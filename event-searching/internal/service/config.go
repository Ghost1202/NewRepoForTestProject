package service

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	JWTSecret            string `envconfig:"JWT_SECRET" default:""`
	CommentsDefaultLimit int64  `envconfig:"COMMENTS_DEFAULT_LIMIT" default:"20"`
	CommentsMaxLimit     int64  `envconfig:"COMMENTS_MAX_LIMIT" default:"100"`
	CommentsMaxOffset    int64  `envconfig:"COMMENTS_MAX_OFFSET" default:"1000"`
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.JWTSecret, validation.Required),

		validation.Field(&cfg.CommentsDefaultLimit, validation.Required),
		validation.Field(&cfg.CommentsMaxLimit, validation.Required),
		validation.Field(&cfg.CommentsMaxOffset, validation.Required),
	)
}
