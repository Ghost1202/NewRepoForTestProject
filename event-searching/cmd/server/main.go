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

	_ "github.com/turtlepavlo/event-searching/docs"

	"github.com/turtlepavlo/event-searching/internal/lib/jwt"
	"github.com/turtlepavlo/event-searching/internal/service"
	"github.com/turtlepavlo/event-searching/internal/storage"
	"github.com/turtlepavlo/event-searching/internal/storage/database/elastic"
	"github.com/turtlepavlo/event-searching/internal/storage/database/postgres"
	"github.com/turtlepavlo/event-searching/internal/storage/database/redis"
	sqlc "github.com/turtlepavlo/event-searching/internal/storage/postgres/sqlc"
	"github.com/turtlepavlo/event-searching/internal/transport/http/handler"

	"github.com/turtlepavlo/sso/pkg/logger"
	"github.com/turtlepavlo/sso/pkg/telemetry"
)

// @title           Event Searching Service API
// @version         1.0
// @description     API service for searching and filtering events using Elasticsearch and Redis.
// @contact.name    API Support
// @contact.email   support@turtlepavlo.com
// @host            localhost:4541
// @BasePath        /api/v1
func main() {
	ctx, stopNotification := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	cfg, err := Load(ctx)
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	zapLogger, shutdownLogger, err := logger.New(cfg.Logger)
	if err != nil {
		log.Fatalf("logger init failed: %v", err)
	}

	defer stopNotification()

	defer func() {
		shutdownContext, cancelLogger := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelLogger()
		if err := shutdownLogger(shutdownContext); err != nil && !errors.Is(err, syscall.ENOTTY) {
			zapLogger.Fatal("logger shutdown error", zap.Error(err))
		}
	}()

	shutdownTracing, err := telemetry.InitTracing(ctx, cfg.Telemetry)
	if err != nil {
		zapLogger.Fatal("tracing init failed", zap.Error(err))
	}
	defer func() {
		shutdownTracingContext, cancelTracing := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelTracing()
		if err := shutdownTracing(shutdownTracingContext); err != nil {
			zapLogger.Warn("tracing shutdown error", zap.Error(err))
		}
	}()

	redisClient, err := redis.New(ctx, cfg.Redis, zapLogger)
	if err != nil {
		zapLogger.Fatal("redis connection failed", zap.Error(err))
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			zapLogger.Warn("redis close failed", zap.Error(err))
		}
	}()

	elasticClient, err := elastic.New(ctx, cfg.Elastic, zapLogger)
	if err != nil {
		zapLogger.Fatal("elastic connection failed", zap.Error(err))
	}

	postgresDB, err := postgres.New(ctx, cfg.Postgres, zapLogger)
	if err != nil {
		zapLogger.Fatal("postgres connection failed", zap.Error(err))
	}
	defer postgresDB.Close()

	redisRepo := storage.NewSearchRepo(
		redisClient,
		zapLogger.With(zap.String("layer", "repository"), zap.String("db", "redis")),
		cfg.Redis.StitchTTL,
	)

	elasticRepo := storage.NewFavouriteStore(
		elasticClient,
		zapLogger.With(zap.String("layer", "repository"), zap.String("db", "elastic")),
	)

	commentRepo := sqlc.New(postgresDB.Pool)

	searchService := service.NewSearchService(
		cfg.Service,
		redisRepo,
		elasticRepo,
		commentRepo,
		zapLogger.With(zap.String("layer", "service")),
	)

	searchHandler := handler.New(
		zapLogger.With(zap.String("layer", "handler")),
		searchService,
		handler.NewRequestConverter(),
		handler.NewResponseConverter(),
	)

	jwtValidator := jwt.NewJWTValidator(cfg.Service.JWTSecret, zapLogger)

	router := handler.NewRouter(
		searchHandler,
		cfg.HTTP,
		zapLogger,
		jwtValidator,
	)

	serverAddress := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
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
		zapLogger.Info("http server starting", zap.String("addr", httpServer.Addr))
		if serveError := httpServer.ListenAndServe(); serveError != nil && !errors.Is(serveError, http.ErrServerClosed) {
			zapLogger.Error("listen and serve failed", zap.Error(serveError))
			stopNotification()
		}
	}()

	<-ctx.Done()
	zapLogger.Info("shutdown signal received")

	gracefulShutdownContext, cancelGraceful := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelGraceful()

	if shutdownError := httpServer.Shutdown(gracefulShutdownContext); shutdownError != nil {
		zapLogger.Error("server graceful shutdown error", zap.Error(shutdownError))
		return
	}

	zapLogger.Info("application stopped gracefully")
}
