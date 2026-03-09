package elastic

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Host     string `envconfig:"HOST_ELASTIC" default:"localhost"`
	Port     int    `envconfig:"PORT_ELASTIC" default:"9200"`
	User     string `envconfig:"USER_ELASTIC" default:""`
	Password string `envconfig:"PASSWORD_ELASTIC" default:""`

	HealthcheckEnabled bool          `envconfig:"HEALTHCHECK_ELASTIC" default:"true"`
	SniffEnabled       bool          `envconfig:"SNIFF_ELASTIC" default:"false"`
	DialTimeout        time.Duration `envconfig:"DIAL_TIMEOUT_ELASTIC" default:"10s"`
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.Host, validation.Required),
		validation.Field(&cfg.Port, validation.Required),
	)
}
