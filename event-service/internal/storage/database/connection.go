package database

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/opentelemetry/tracing"
)

type Database struct {
	gormConnection *gorm.DB
}

func New(ctx context.Context, config Config, zapLogger *zap.Logger) (*Database, error) {
	const operation = "infrastructure.postgres.New"

	log := zapLogger.With(
		zap.String("layer", "infrastructure"),
		zap.String("component", "postgres"),
		zap.String("op", operation),
	)

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d application_name=%s statement_timeout=%d",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
		int(config.ConnectTimeout.Seconds()),
		config.AppName,
		int(config.StatementTimeout.Milliseconds()),
	)

	gormConnection, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Silent),
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		log.Error("failed to open gorm connection", zap.Error(err))
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
		log.Error("failed to ping database",
			zap.String("db.host", config.Host),
			zap.Int("db.port", config.Port),
			zap.Error(err),
		)
		return nil, err
	}

	log.Info("successfully connected to postgres",
		zap.String("db.host", config.Host),
		zap.Int("db.port", config.Port),
		zap.String("db.name", config.DBName),
		zap.String("db.app", config.AppName),
	)

	return &Database{gormConnection: gormConnection}, nil
}

func (database *Database) Close() error {
	if database == nil || database.gormConnection == nil {
		return nil
	}
	sqlConnection, err := database.gormConnection.DB()
	if err != nil {
		return err
	}
	return sqlConnection.Close()
}

func (database *Database) DB() *gorm.DB {
	if database == nil {
		return nil
	}
	return database.gormConnection
}
