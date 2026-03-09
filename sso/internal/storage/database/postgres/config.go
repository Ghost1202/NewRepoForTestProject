package postgres

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Host     string `envconfig:"DB_HOST" default:"localhost"`
	Port     int    `envconfig:"DB_PORT" default:"5432"`
	User     string `envconfig:"DB_USER" default:"postgres"`
	Password string `envconfig:"DB_PASSWORD" default:"postgres"`
	DBName   string `envconfig:"DB_NAME" default:"sso"`
	SSLMode  string `envconfig:"DB_SSLMODE" default:"disable"`

	AppName          string        `envconfig:"APP_NAME" default:"sso"`
	ConnectTimeout   time.Duration `envconfig:"DB_CONNECT_TIMEOUT" default:"10s"`
	StatementTimeout time.Duration `envconfig:"DB_STATEMENT_TIMEOUT" default:"30s"`

	MaxOpenConns    int           `envconfig:"DB_MAX_OPEN_CONNS" default:"10"`
	MaxIdleConns    int           `envconfig:"DB_MAX_IDLE_CONNS" default:"5"`
	ConnMaxLifetime time.Duration `envconfig:"DB_CONN_MAX_LIFETIME" default:"5m"`
	ConnMaxIdleTime time.Duration `envconfig:"DB_CONN_MAX_IDLE_TIME" default:"10m"`
}

func (cfg Config) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.Host, validation.Required),
		validation.Field(&cfg.User, validation.Required),
		validation.Field(&cfg.Password, validation.Required),
		validation.Field(&cfg.DBName, validation.Required),

		validation.Field(&cfg.ConnectTimeout, validation.Required),

		validation.Field(&cfg.MaxOpenConns, validation.Required),
		validation.Field(&cfg.ConnMaxLifetime, validation.Required),
		validation.Field(&cfg.ConnMaxIdleTime, validation.Required),
	)
}
