package jwt

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	JWTSecret string `envconfig:"JWT_SECRET" required:"true"`
}

func (cfg Config) Validate() error {
	return validation.ValidateStruct(&cfg,
		validation.Field(&cfg.JWTSecret, validation.Required),
	)
}
