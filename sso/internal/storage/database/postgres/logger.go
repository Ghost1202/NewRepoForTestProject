package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ZapGormLogger struct {
	ZapLogger *zap.Logger
	LogLevel  logger.LogLevel
}

func (log *ZapGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	log.LogLevel = level
	return log
}

func (log *ZapGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if log == nil || log.ZapLogger == nil {
		return
	}
	if log.LogLevel < logger.Info {
		return
	}
	log.ZapLogger.Info(fmt.Sprintf(msg, data...), traceFields(ctx)...)
}

func (log *ZapGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if log == nil || log.ZapLogger == nil {
		return
	}
	if log.LogLevel < logger.Warn {
		return
	}
	log.ZapLogger.Warn(fmt.Sprintf(msg, data...), traceFields(ctx)...)
}

func (log *ZapGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if log == nil || log.ZapLogger == nil {
		return
	}
	if log.LogLevel < logger.Error {
		return
	}
	log.ZapLogger.Error(fmt.Sprintf(msg, data...), traceFields(ctx)...)
}

func (log *ZapGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if log == nil || log.ZapLogger == nil {
		return
	}
	if log.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sqlQuery, rows := fc()

	fields := []zap.Field{
		zap.String("db.system", "postgresql"),
		zap.String("db.operation", "SQL"),
		zap.Duration("duration", elapsed),
		zap.Int64("rows", rows),
		zap.String("sql", sqlQuery),
	}
	fields = append(fields, traceFields(ctx)...)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.ZapLogger.Info("db query (record not found)", append(fields,
				zap.String("reason", "record_not_found"),
			)...)
			return
		}

		log.ZapLogger.Error("db query failed", append(fields,
			zap.Error(err),
		)...)
		return
	}

	log.ZapLogger.Info("db query", fields...)
}

func traceFields(ctx context.Context) []zap.Field {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return nil
	}
	return []zap.Field{
		zap.String("trace_id", sc.TraceID().String()),
		zap.String("span_id", sc.SpanID().String()),
	}
}
