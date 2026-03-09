package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	_ "github.com/turtlepavlo/sso/internal/docs"
	"github.com/turtlepavlo/sso/internal/service"
	"github.com/turtlepavlo/sso/internal/storage"
	"github.com/turtlepavlo/sso/internal/storage/database/postgres"
	"github.com/turtlepavlo/sso/internal/storage/database/redis"
	httptransport "github.com/turtlepavlo/sso/internal/transport/http/handler"
	"github.com/turtlepavlo/sso/pkg/logger"
	"github.com/turtlepavlo/sso/pkg/producer"
	"github.com/turtlepavlo/sso/pkg/provider/google"
	"github.com/turtlepavlo/sso/pkg/provider/infobip"
	"github.com/turtlepavlo/sso/pkg/provider/resend"
	"github.com/turtlepavlo/sso/pkg/telemetry"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	cfg, err := Load(ctx)
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	zapLogger, shutdownLogger, err := logger.New(cfg.Logger)
	if err != nil {
		log.Fatalf("logger init failed: %v", err)
	}
	defer stop()
	defer func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := shutdownLogger(c); err != nil && !errors.Is(err, syscall.ENOTTY) {
			zapLogger.Warn("logger shutdown error", zap.Error(err))
		}
	}()

	appLogger := zapLogger.With(zap.String("component", "sso"))
	httpLogger := zapLogger.With(zap.String("component", "http"))
	dbLogger := zapLogger.With(zap.String("component", "db"))

	serverAddress := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)

	analyticsProducer := producer.New(cfg.Producer)
	defer func() {
		if closeErr := analyticsProducer.Close(); closeErr != nil {
			appLogger.Warn("failed to close analytics producer", zap.Error(closeErr))
		}
	}()

	shutdownTracing, err := telemetry.InitTracing(ctx, cfg.Telemetry)
	if err != nil {
		appLogger.Fatal("tracing init failed", zap.Error(err))
	}
	defer func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := shutdownTracing(c); err != nil {
			appLogger.Warn("tracing shutdown error", zap.Error(err))
		}
	}()

	db, err := postgres.New(ctx, cfg.DB, dbLogger)
	if err != nil {
		dbLogger.Fatal("database connection failed", zap.Error(err))
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			dbLogger.Warn("database close failed", zap.Error(closeErr))
		}
	}()

	cache, err := redis.New(ctx, cfg.Cache, dbLogger)
	if err != nil {
		appLogger.Fatal("cache connection failed", zap.Error(err))
	}
	defer func() {
		if closeErr := cache.Close(); closeErr != nil {
			appLogger.Warn("cache close failed", zap.Error(closeErr))
		}
	}()

	userRepo := storage.NewUserRepo(db.DB(), dbLogger)
	cacheRepo := storage.NewCacheRepo(cache, dbLogger, cfg.Cache.LockTTL)

	googleProvider := google.NewClient(cfg.Google, appLogger)
	smsProvider := infobip.NewClient(cfg.SMSProvider, appLogger)

	emailProvider := resend.NewClient(cfg.EmeilProvider, appLogger)

	authService := service.New(
		userRepo,
		cacheRepo,
		appLogger,
		cfg.JWT.JWTSecret,
		cfg.JWT.TokenTTL,
		analyticsProducer,
		googleProvider,
		smsProvider,
		emailProvider,
		cfg.Service.OTPMin,
		cfg.Service.OTPMax,
	)

	authHandler := httptransport.New(authService, httpLogger, cfg.HTTP.FrontendURL)

	router := httptransport.NewRouter(
		authHandler,
		cfg.Middleware,
		httpLogger,
		cfg.JWT.JWTSecret,
		cfg.Telemetry.ServiceName,
	)

	httpServer := &http.Server{
		Addr:              serverAddress,
		Handler:           router,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
		MaxHeaderBytes:    cfg.HTTP.MaxHeaderBytes,
	}

	go func() {
		httpLogger.Info("http server starting", zap.String("addr", httpServer.Addr))
		if serveErr := httpServer.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			httpLogger.Error("listen and serve failed", zap.Error(serveErr))
			stop()
		}
	}()

	<-ctx.Done()
	httpLogger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if shutdownErr := httpServer.Shutdown(shutdownCtx); shutdownErr != nil {
		httpLogger.Error("server graceful shutdown error", zap.Error(shutdownErr))
		return
	}

	appLogger.Info("application stopped gracefully")
}
