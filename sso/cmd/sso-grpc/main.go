package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	grpcapp "github.com/turtlepavlo/sso/internal/app/grpc"
	"github.com/turtlepavlo/sso/internal/service"
	"github.com/turtlepavlo/sso/internal/storage"
	"github.com/turtlepavlo/sso/internal/storage/database/postgres"
	"github.com/turtlepavlo/sso/internal/storage/database/redis"
	"github.com/turtlepavlo/sso/internal/transport/grpc/server"
	"github.com/turtlepavlo/sso/pkg/logger"
	"github.com/turtlepavlo/sso/pkg/producer"
	"github.com/turtlepavlo/sso/pkg/provider/google"
	"github.com/turtlepavlo/sso/pkg/provider/infobip"
	"github.com/turtlepavlo/sso/pkg/provider/resend"
	"github.com/turtlepavlo/sso/pkg/telemetry"
)

func main() {
	ctx, stopNotification := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	config, err := Load(ctx)
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	logger, shutdownLogger, err := logger.New(config.Logger)
	if err != nil {
		log.Fatalf("logger init failed: %v", err)
	}

	defer stopNotification()

	defer func() {
		if err := shutdownLogger(context.Background()); err != nil {
			log.Fatalf("failed to shutdown logger: %v", err)
		}
	}()

	producer := producer.New(config.Producer)
	defer func() {
		if closeErr := producer.Close(); closeErr != nil {
			logger.Fatal("failed to close analytics producer", zap.Error(closeErr))
		}
	}()

	shutdownTracing, err := telemetry.InitTracing(ctx, config.Telemetry)
	if err != nil {
		logger.Fatal("tracing init failed", zap.Error(err))
	}
	defer func() {
		if err := shutdownTracing(context.Background()); err != nil {
			logger.Error("failed to shutdown tracing", zap.Error(err))
		}
	}()
	db, err := postgres.New(ctx, config.DB, logger)
	if err != nil {
		logger.Fatal("database connection failed", zap.Error(err))
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			logger.Warn("database close failed", zap.Error(closeErr))
		}
	}()
	cache, err := redis.New(ctx, config.Cache, logger)
	if err != nil {
		logger.Fatal("cache connection failed", zap.Error(err))
	}
	defer func() {
		if closeError := cache.Close(); closeError != nil {
			logger.Fatal("database close failed", zap.Error(closeError))
		}
	}()

	userRepo := storage.NewUserRepo(db.DB(), logger)
	cacheRepo := storage.NewCacheRepo(cache, logger, config.Cache.LockTTL)

	googleProvider := google.NewClient(config.Google, logger)
	smsProvider := infobip.NewClient(config.SMSProvider, logger)

	emailProvider := resend.NewClient(config.EmeilProvider, logger)

	authService := service.New(
		userRepo,
		cacheRepo,
		logger,
		config.JWT.JWTSecret,
		config.JWT.TokenTTL,
		producer,
		googleProvider,
		smsProvider,
		emailProvider,
		config.Service.OTPMin,
		config.Service.OTPMax,
	)

	grpcHandler := server.New(
		authService,
		config.JWT.JWTSecret,
		logger,
	)
	application := grpcapp.New(
		logger,
		grpcHandler,
		config.GRPCServer,
		config.GRPCInterceptor,
		config.JWT.JWTSecret,
	)
	go func() {
		if err := application.Run(); err != nil {
			logger.Fatal("application run failed", zap.Error(err))
			stopNotification()
		}
	}()

	logger.Info("application started",
		zap.String("addr", config.GRPCServer.Host),
		zap.Int("port", config.GRPCServer.Port),
	)
	<-ctx.Done()

	logger.Info("shutdown signal received")
	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	stopDoneChannel := make(chan struct{})
	go func() {
		application.Stop()
		close(stopDoneChannel)
	}()

	select {
	case <-stopDoneChannel:
		logger.Info("application stopped gracefully")
	case <-shutdownContext.Done():
		logger.Warn("shutdown timeout reached, stopping forcefully")
	}
}
