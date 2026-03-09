package redis

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Host     string `envconfig:"HOST_REDIS" default:"localhost"`
	Port     int    `envconfig:"PORT_REDIS" default:"2315"`
	Password string `envconfig:"PASSWORD_REDIS" default:""`
	DB       int    `envconfig:"DB_REDIS" default:"0"`

	DialTimeout  time.Duration `envconfig:"DIAL_TIMEOUT_REDIS" default:"5s"`
	ReadTimeout  time.Duration `envconfig:"READ_TIMEOUT_REDIS" default:"3s"`
	WriteTimeout time.Duration `envconfig:"WRITE_TIMEOUT_REDIS" default:"3s"`
	PoolSize     int           `envconfig:"POOL_SIZE_REDIS" default:"10"`
	MinIdleConns int           `envconfig:"MIN_IDLE_CONNS_REDIS" default:"5"`
	LockTTL      time.Duration `envconfig:"LOCK_TTL_REDIS" default:"10m"`
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.Host, validation.Required),
		validation.Field(&cfg.Port, validation.Required),
	)
}
