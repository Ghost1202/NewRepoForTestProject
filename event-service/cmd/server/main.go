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

	_ "github.com/turtlepavlo/event-service/docs"
	"github.com/turtlepavlo/event-service/internal/lib/jwt"
	"github.com/turtlepavlo/event-service/internal/service"
	"github.com/turtlepavlo/event-service/internal/storage"
	"github.com/turtlepavlo/event-service/internal/storage/database"
	payment "github.com/turtlepavlo/event-service/internal/transport/consumer/payment"
	"github.com/turtlepavlo/event-service/internal/transport/consumer/transfer"
	"github.com/turtlepavlo/event-service/internal/transport/consumer/wallet"
	"github.com/turtlepavlo/event-service/internal/transport/http/handler"
	"github.com/turtlepavlo/sso/pkg/logger"
	"github.com/turtlepavlo/sso/pkg/telemetry"
)

// @title Event Service API
// @version 1.0
// @description API for managing events, sectors, and ticket configurations.
// @securityDefinitions.apiKey ApiKeyAuth
// @in header
// @name Authorization

// @host localhost:6660
// @BasePath /api/v1
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
		shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()
		if err := shutdownLogger(shutdownContext); err != nil && !errors.Is(err, syscall.ENOTTY) {
			zapLogger.Warn("logger shutdown error", zap.Error(err))
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

	db, err := database.New(ctx, cfg.DB, zapLogger)
	if err != nil {
		zapLogger.Fatal("database connection failed", zap.Error(err))
	}
	defer func() {
		if closeError := db.Close(); closeError != nil {
			zapLogger.Warn("database close failed", zap.Error(closeError))
		}
	}()

	repository := storage.NewRepository(
		db.DB(),
		zapLogger,
	)

	eventService := service.NewEventService(
		repository,
		zapLogger,
	)

	jwtValidator := jwt.NewJWTValidator(
		cfg.Transport.JWTSecret,
		zapLogger,
	)

	paymentKafkaBrokerAddr := fmt.Sprintf("%s:%d", cfg.PaymentConsumer.Host, cfg.PaymentConsumer.Port)

	paymentConsumer := payment.NewConsumer(
		[]string{paymentKafkaBrokerAddr},
		cfg.PaymentConsumer,
		zapLogger,
		eventService,
	)

	walletKafkaBrokerAddr := fmt.Sprintf("%s:%d", cfg.WalletConsumer.Host, cfg.WalletConsumer.Port)

	walletConsumer := wallet.NewConsumer(
		[]string{walletKafkaBrokerAddr},
		cfg.WalletConsumer,
		zapLogger,
		eventService,
	)

	transferKafkaBrokerAddr := fmt.Sprintf("%s:%d", cfg.TransferConsumer.Host, cfg.TransferConsumer.Port)

	transferConsumer := transfer.NewConsumer(
		[]string{transferKafkaBrokerAddr},
		cfg.TransferConsumer,
		zapLogger,
		eventService,
	)
	go transferConsumer.Start(ctx)
	defer func() {
		if err := transferConsumer.Close(); err != nil {
			zapLogger.Error("kafka consumer close failed", zap.Error(err))
		}
	}()
	go paymentConsumer.Start(ctx)

	defer func() {
		if err := paymentConsumer.Close(); err != nil {
			zapLogger.Error("kafka consumer close failed", zap.Error(err))
		}
	}()

	go walletConsumer.Start(ctx)
	defer func() {
		if err := walletConsumer.Close(); err != nil {
			zapLogger.Error("kafka consumer close failed", zap.Error(err))
		}
	}()

	eventHandler := handler.New(
		eventService,
		zapLogger,
		handler.NewRequestConverter(),
		handler.NewResponseConverter(),
	)

	router := handler.NewRouter(
		eventHandler,
		cfg.Transport,
		zapLogger,
		jwtValidator,
	)

	serverAddress := fmt.Sprintf("%s:%d", cfg.Transport.Host, cfg.Transport.Port)
	httpServer := &http.Server{
		Addr:              serverAddress,
		Handler:           router,
		ReadTimeout:       cfg.Transport.ReadTimeout,
		ReadHeaderTimeout: cfg.Transport.ReadHeaderTimeout,
		WriteTimeout:      cfg.Transport.WriteTimeout,
		IdleTimeout:       cfg.Transport.IdleTimeout,
		MaxHeaderBytes:    cfg.Transport.MaxHeaderBytes,
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
