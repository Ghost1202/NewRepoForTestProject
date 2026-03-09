package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	gogrpc "google.golang.org/grpc"

	paymentv1 "github.com/turtlepavlo/proto-contract/gen/go/payment/v1"
	"github.com/turtlepavlo/sso/pkg/logger"
	"github.com/turtlepavlo/sso/pkg/telemetry"

	"github.com/turtlepavlo/stripe_integration/internal/service"

	dbConn "github.com/turtlepavlo/stripe_integration/internal/storage/database/postgres"
	storagePg "github.com/turtlepavlo/stripe_integration/internal/storage/postgres"

	grpcServerPkg "github.com/turtlepavlo/stripe_integration/internal/transport/grpc/server"
	handler "github.com/turtlepavlo/stripe_integration/internal/transport/http/handler"
	"github.com/turtlepavlo/stripe_integration/pkg/producer"
	"github.com/turtlepavlo/stripe_integration/pkg/provider/resend"
	stripe "github.com/turtlepavlo/stripe_integration/pkg/stripe"
)

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
			zapLogger.Warn("logger shutdown failed", zap.Error(err))
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

	pgDB, err := dbConn.New(ctx, cfg.Postgres, zapLogger)
	if err != nil {
		zapLogger.Fatal("database connection failed", zap.Error(err))
	}
	defer pgDB.Close()

	kafkaProducer := producer.New(cfg.Producer, zapLogger)
	defer kafkaProducer.Close()

	stripeClient := stripe.New(cfg.Stripe, zapLogger)
	resendClient := resend.NewClient(cfg.EmeilProvider, zapLogger)
	paymentRepo := storagePg.NewRepository(pgDB.Pool)
	paymentService := service.NewPaymentService(
		cfg.Service,
		paymentRepo,
		stripeClient,
		kafkaProducer,
		resendClient,
		zapLogger,
	)

	httpHandler := handler.NewHandler(zapLogger, paymentService, cfg.HTTP)
	httpRouter := handler.NewRouter(httpHandler, cfg.HTTP, zapLogger)

	httpServerAddress := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	httpServer := &http.Server{
		Addr:           httpServerAddress,
		Handler:        httpRouter,
		ReadTimeout:    cfg.HTTP.ReadTimeout,
		WriteTimeout:   cfg.HTTP.WriteTimeout,
		IdleTimeout:    cfg.HTTP.IdleTimeout,
		MaxHeaderBytes: cfg.HTTP.MaxHeaderBytes,
	}

	go func() {
		zapLogger.Info("http server starting", zap.String("addr", httpServerAddress))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zapLogger.Error("http server failed", zap.Error(err))
			stopNotification()
		}
	}()

	grpcListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port))
	if err != nil {
		zapLogger.Error("failed to listen grpc port", zap.Error(err))
		return
	}

	serverImplementation := grpcServerPkg.New(paymentService, zapLogger)
	baseGrpcServer := gogrpc.NewServer()
	paymentv1.RegisterPaymentServiceServer(baseGrpcServer, serverImplementation)

	go func() {
		zapLogger.Info("grpc server starting", zap.String("addr", grpcListener.Addr().String()))
		if err := baseGrpcServer.Serve(grpcListener); err != nil && !errors.Is(err, gogrpc.ErrServerStopped) {
			zapLogger.Error("grpc server failed", zap.Error(err))
			stopNotification()
		}
	}()

	<-ctx.Done()
	zapLogger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		zapLogger.Error("http graceful shutdown failed", zap.Error(err))
	} else {
		zapLogger.Info("http server stopped")
	}

	baseGrpcServer.GracefulStop()
	zapLogger.Info("grpc server stopped")

	zapLogger.Info("application stopped gracefully")
}
