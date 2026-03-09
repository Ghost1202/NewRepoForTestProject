package postgres

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Host     string `envconfig:"EVENT_SEARCHING_DB_HOST" default:"localhost"`
	Port     int    `envconfig:"EVENT_SEARCHING_DB_PORT" default:"5432"`
	User     string `envconfig:"EVENT_SEARCHING_DB_USER" default:"postgres"`
	Password string `envconfig:"EVENT_SEARCHING_DB_PASSWORD" default:"postgres"`
	DBName   string `envconfig:"EVENT_SEARCHING_DB_NAME" default:"event_searching_db"`
	SSLMode  string `envconfig:"EVENT_SEARCHING_DB_SSLMODE" default:"disable"`

	AppName          string        `envconfig:"EVENT_SEARCHING_APP_NAME" default:"event_searching"`
	ConnectTimeout   time.Duration `envconfig:"EVENT_SEARCHING_DB_CONNECT_TIMEOUT" default:"10s"`
	StatementTimeout time.Duration `envconfig:"EVENT_SEARCHING_DB_STATEMENT_TIMEOUT" default:"30s"`

	MaxOpenConns    int32         `envconfig:"EVENT_SEARCHING_DB_MAX_OPEN_CONNS" default:"10"`
	MaxIdleConns    int32         `envconfig:"EVENT_SEARCHING_DB_MAX_IDLE_CONNS" default:"5"`
	ConnMaxLifetime time.Duration `envconfig:"EVENT_SEARCHING_DB_CONN_MAX_LIFETIME" default:"5m"`
	ConnMaxIdleTime time.Duration `envconfig:"EVENT_SEARCHING_DB_CONN_MAX_IDLE_TIME" default:"10m"`
}

func (cfg Config) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.Host, validation.Required),
		validation.Field(&cfg.Port, validation.Required),
		validation.Field(&cfg.User, validation.Required),
		validation.Field(&cfg.Password, validation.Required),
		validation.Field(&cfg.DBName, validation.Required),

		validation.Field(&cfg.ConnectTimeout, validation.Required),
		validation.Field(&cfg.StatementTimeout, validation.Required),

		validation.Field(&cfg.MaxOpenConns, validation.Required),
		validation.Field(&cfg.MaxIdleConns, validation.Required),
		validation.Field(&cfg.ConnMaxLifetime, validation.Required),
		validation.Field(&cfg.ConnMaxIdleTime, validation.Required),
	)
}
