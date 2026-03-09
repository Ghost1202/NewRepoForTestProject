package postgres

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
)

type Database struct {
	gormConnection *gorm.DB
}

func New(ctx context.Context, config Config, zapLogger *zap.Logger) (*Database, error) {
	const operation = "event-booking/internal/storage/database/postgres.New"

	log := zapLogger.With(
		zap.String("layer", "storage"),
		zap.String("component", "postgres-slave"),
		zap.String("op", operation),
	)

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d application_name=%s statement_timeout=%s",
		config.ReplicaHost,
		config.ReplicaPort,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
		config.ConnectTimeout/time.Second,
		config.AppName,
		config.StatementTimeout.String(),
	)

	gormConnection, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Error("failed to open gorm connection to slave", zap.Error(err))
		return nil, err
	}

	if err := gormConnection.Use(tracing.NewPlugin(tracing.WithoutQueryVariables())); err != nil {
		log.Error("failed to init gorm otel plugin", zap.Error(err))
		return nil, err
	}

	sqlConnection, err := gormConnection.DB()
	if err != nil {
		log.Error("failed to get sql connection from gorm", zap.Error(err))
		return nil, err
	}

	sqlConnection.SetMaxOpenConns(config.MaxOpenConns)
	sqlConnection.SetMaxIdleConns(config.MaxIdleConns)
	sqlConnection.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlConnection.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	if err := sqlConnection.PingContext(ctx); err != nil {
		log.Error("failed to ping slave database",
			zap.String("db.host", config.ReplicaHost),
			zap.Int64("db.port", int64(config.ReplicaPort)),
			zap.Error(err),
		)
		return nil, err
	}

	log.Info("successfully connected to postgres slave",
		zap.String("db.host", config.ReplicaHost),
		zap.Int64("db.port", int64(config.ReplicaPort)),
		zap.String("db.name", config.DBName),
		zap.String("db.role", "slave"),
	)

	return &Database{gormConnection: gormConnection}, nil
}

func (d *Database) DB() *gorm.DB {
	return d.gormConnection
}

func (d *Database) Close() error {
	sqlDB, err := d.gormConnection.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
