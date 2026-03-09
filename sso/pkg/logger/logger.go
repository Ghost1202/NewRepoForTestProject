package logger

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

type ShutdownFunc func(ctx context.Context) error

func New(conf Config) (*zap.Logger, ShutdownFunc, error) {
	if err := conf.ValidateWithContext(context.Background()); err != nil {
		return nil, nil, err
	}

	zcfg, err := buildZapConfig(conf)
	if err != nil {
		return nil, nil, err
	}

	var (
		lp          *sdklog.LoggerProvider
		otelCore    zapcore.Core
		shutdownOTL ShutdownFunc = func(context.Context) error { return nil }
	)

	if conf.OTLPEnabled {
		initCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		exp, err := newOTLPLogExporter(initCtx, conf)
		if err != nil {
			return nil, nil, err
		}

		res, err := resource.New(initCtx,
			resource.WithFromEnv(),
			resource.WithTelemetrySDK(),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("otlp log resource init: %w", err)
		}

		lp = sdklog.NewLoggerProvider(
			sdklog.WithResource(res),
			sdklog.WithProcessor(sdklog.NewBatchProcessor(exp)),
		)

		otelCore = otelzap.NewCore(
			conf.OTLPScope,
			otelzap.WithLoggerProvider(lp),
		)

		shutdownOTL = func(ctx context.Context) error {
			return lp.Shutdown(ctx)
		}
	}

	log, err := zcfg.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			if !conf.OTLPEnabled || otelCore == nil {
				return core
			}
			return zapcore.NewTee(core, otelCore)
		}),
	)
	if err != nil {
		_ = shutdownOTL(context.Background())
		return nil, nil, err
	}

	shutdown := func(ctx context.Context) error {
		if err := shutdownOTL(ctx); err != nil {
			return err
		}

		if err := log.Sync(); err != nil && !errors.Is(err, syscall.ENOTTY) {
			return err
		}
		return nil
	}

	return log, shutdown, nil
}

func buildZapConfig(conf Config) (zap.Config, error) {
	var cfg zap.Config

	mode := strings.ToLower(conf.Mode)
	if mode == "dev" || mode == "development" {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	var lvl zapcore.Level
	if err := lvl.Set(strings.ToLower(conf.Level)); err != nil {
		return zap.Config{}, fmt.Errorf("invalid log level %q: %w", conf.Level, err)
	}
	cfg.Level = zap.NewAtomicLevelAt(lvl)

	cfg.Encoding = conf.Encoding
	cfg.OutputPaths = []string{conf.Output}
	cfg.ErrorOutputPaths = []string{conf.ErrorOutput}

	cfg.EncoderConfig.TimeKey = conf.TimeKey
	cfg.EncoderConfig.LevelKey = conf.LevelKey
	cfg.EncoderConfig.MessageKey = conf.MessageKey
	cfg.EncoderConfig.CallerKey = conf.CallerKey

	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	return cfg, nil
}

func newOTLPLogExporter(ctx context.Context, conf Config) (*otlploggrpc.Exporter, error) {
	opts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(conf.OTLPEndpoint),
	}
	if conf.Insecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}

	exp, err := otlploggrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("otlp log exporter init: %w", err)
	}
	return exp, nil
}
