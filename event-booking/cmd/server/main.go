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

	_ "github.com/turtlepavlo/event-booking/docs"
	"github.com/turtlepavlo/event-booking/internal/lib/jwt"
	"github.com/turtlepavlo/event-booking/internal/service"
	"github.com/turtlepavlo/event-booking/internal/storage"
	"github.com/turtlepavlo/event-booking/internal/storage/database/postgres"
	"github.com/turtlepavlo/event-booking/internal/storage/database/redis"
	paymentclient "github.com/turtlepavlo/event-booking/internal/transport/client/payment"
	"github.com/turtlepavlo/event-booking/internal/transport/client/wallet"
	"github.com/turtlepavlo/event-booking/internal/transport/http/handler"
	"github.com/turtlepavlo/event-booking/pkg/producer"
	"github.com/turtlepavlo/sso/pkg/logger"
	"github.com/turtlepavlo/sso/pkg/telemetry"
)

// @title           Event Booking Service API
// @version         1.0
// @description     API service for event ticket booking. Features include Redis-based ticket locking, Postgres slave-read optimization, and Kafka event streaming for payments.
// @contact.name    POP
// @contact.url     https://github.com/turtlepavlo

// @host      localhost:3434
// @BasePath  /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type 'Bearer ' followed by your JWT token.
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
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownLogger(shutdownCtx); err != nil {
			zapLogger.Fatal("logger shutdown failed", zap.Error(err))
		}
	}()

	shutdownTracing, err := telemetry.InitTracing(ctx, cfg.Telemetry)
	if err != nil {
		zapLogger.Fatal("tracing init failed", zap.Error(err))
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTracing(shutdownCtx); err != nil {
			zapLogger.Warn("tracing shutdown failed", zap.Error(err))
		}
	}()

	pgDB, err := postgres.New(ctx, cfg.Postgres, zapLogger)
	if err != nil {
		zapLogger.Fatal("database connection failed", zap.Error(err))
	}
	defer pgDB.Close()

	redisDB, err := redis.New(ctx, cfg.Redis, zapLogger)
	if err != nil {
		zapLogger.Fatal("redis connection failed", zap.Error(err))
	}
	defer redisDB.Close()

	paymentAddr := fmt.Sprintf("%s:%d", cfg.PaymentClient.Host, cfg.PaymentClient.Port)
	paymentClient, err := paymentclient.NewPaymentClient(paymentAddr, cfg.PaymentClient.TokenTTL, zapLogger)
	if err != nil {
		zapLogger.Fatal("payment client init failed", zap.Error(err))
	}
	defer paymentClient.Close()

	walletClient, err := wallet.NewWalletClient(cfg.WalletClient, zapLogger)
	if err != nil {
		zapLogger.Fatal("wallet client init failed", zap.Error(err))
	}
	defer walletClient.Close()

	ticketRepo := storage.NewTicketRepo(pgDB.DB(), zapLogger)
	cacheRepo := storage.NewCacheRepo(redisDB, zapLogger)
	producer := producer.New(cfg.Producer, zapLogger)
	defer func() {
		if err := producer.Close(); err != nil {
			zapLogger.Warn("producer close failed", zap.Error(err))
		}
	}()

	bookingService := service.New(
		ticketRepo,
		cacheRepo,
		paymentClient,
		walletClient,
		zapLogger,
		cfg.Service,
		producer,
	)

	jwtValidator := jwt.NewValidator(cfg.JWT.JWTSecret, zapLogger)

	bookingHandler := handler.New(
		zapLogger,
		bookingService,
		*handler.NewResponseConverter(),
		*handler.NewRequestConverter(),
		cfg.Transport,
	)

	router := handler.NewRouter(
		bookingHandler,
		cfg.Transport,
		zapLogger,
		jwtValidator,
	)

	serverAddress := fmt.Sprintf("%s:%d", cfg.Transport.Host, cfg.Transport.Port)
	httpServer := &http.Server{
		Addr:           serverAddress,
		Handler:        router,
		ReadTimeout:    cfg.Transport.ReadTimeout,
		WriteTimeout:   cfg.Transport.WriteTimeout,
		IdleTimeout:    cfg.Transport.IdleTimeout,
		MaxHeaderBytes: cfg.Transport.MaxHeaderBytes,
	}

	go func() {
		zapLogger.Info("http server starting", zap.String("addr", serverAddress))
		if serveErr := httpServer.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			zapLogger.Error("server failed", zap.Error(serveErr))
			stopNotification()
		}
	}()

	<-ctx.Done()
	zapLogger.Info("shutdown signal received")

	gracefulCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(gracefulCtx); err != nil {
		zapLogger.Error("graceful shutdown error", zap.Error(err))
	}

	zapLogger.Info("application stopped gracefully")
}
