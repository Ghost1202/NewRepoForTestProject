package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/turtlepavlo/event-searching/internal/lib/telemetry"
)

type Database struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, cfg Config, zapLogger *zap.Logger) (*Database, error) {
	const operation = "postgres.New"

	log := telemetry.WithTrace(ctx, zapLogger).With(
		zap.String("layer", "storage"),
		zap.String("component", "postgres"),
		zap.String("op", operation),
	)

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Error("failed to parse postgres config", zap.Error(err))
		return nil, err
	}

	poolConfig.MaxConns = cfg.MaxOpenConns
	poolConfig.MinConns = cfg.MaxIdleConns
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = cfg.ConnMaxIdleTime

	poolConfig.HealthCheckPeriod = 1 * time.Minute
	poolConfig.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	if poolConfig.ConnConfig.RuntimeParams == nil {
		poolConfig.ConnConfig.RuntimeParams = make(map[string]string)
	}
	poolConfig.ConnConfig.RuntimeParams["statement_timeout"] = fmt.Sprintf("%d", cfg.StatementTimeout.Milliseconds())
	poolConfig.ConnConfig.RuntimeParams["application_name"] = cfg.AppName

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Error("failed to connect to postgres", zap.Error(err))
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		log.Error("failed to ping database",
			zap.String("db.host", cfg.Host),
			zap.Int("db.port", cfg.Port),
			zap.String("db.name", cfg.DBName),
			zap.Error(err),
		)
		return nil, err
	}

	log.Info("successfully connected to postgres",
		zap.String("db.host", cfg.Host),
		zap.Int("db.port", cfg.Port),
		zap.String("db.name", cfg.DBName),
	)

	return &Database{Pool: pool}, nil
}

func (d *Database) Close() {
	if d.Pool != nil {
		d.Pool.Close()
	}
}
