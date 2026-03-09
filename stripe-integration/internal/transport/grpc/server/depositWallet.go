package server

import (
	"context"
	"errors"

	grpcCodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	paymentv1 "github.com/turtlepavlo/proto-contract/gen/go/payment/v1"
	"github.com/turtlepavlo/stripe_integration/pkg/telemetry"
)

func (s *Server) DepositWallet(ctx context.Context, req *paymentv1.TopupWalletRequest) (*paymentv1.TopupWalletResponse, error) {
	const op = "server.Server.DepositWallet"

	tracer := otel.Tracer("stripe-integration/internal/transport/grpc/server")
	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.service", "payment.v1.PaymentService"),
			attribute.String("rpc.method", "TopupWallet"),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, s.log).With(zap.String("op", op))

	if req == nil {
		err := status.Error(grpcCodes.InvalidArgument, "request is required")
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, "nil request")
		log.Warn("nil request")
		return nil, err
	}

	span.SetAttributes(
		attribute.String("order_id", req.GetOrderId()),
		attribute.Int64("user_id", req.GetUserId()),
		attribute.Int64("wallet_id", req.GetWalletId()),
		attribute.Int64("amount", req.GetAmount()),
		attribute.String("currency", req.GetCurrency()),
	)

	topup, err := s.reqConv.ToTopupWalletInput(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, "invalid order_id format")
		log.Warn("invalid order_id format", zap.Error(err))
		return nil, status.Error(grpcCodes.InvalidArgument, "invalid order_id format")
	}

	log.Debug("creating topup payment link")
	paymentURL, expiresAt, err := s.service.DepositWallet(ctx, topup)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, "DepositWallet failed")
		switch {
		case errors.Is(err, ErrPaymentAlreadyPaid):
			log.Warn("order already paid", zap.Error(err))
			return nil, status.Error(grpcCodes.FailedPrecondition, "order is already paid")
		case errors.Is(err, ErrPaymentLinkExists):
			log.Warn("payment link already exists", zap.Error(err))
			return nil, status.Error(grpcCodes.AlreadyExists, "payment link already exists")
		case errors.Is(err, ErrInvalidTokenTTL):
			log.Warn("invalid topup request", zap.Error(err))
			return nil, status.Error(grpcCodes.InvalidArgument, err.Error())
		default:
			log.Error("internal error while creating topup link", zap.Error(err))
			return nil, status.Error(grpcCodes.Internal, "internal server error")
		}
	}

	log.Info("topup payment link created successfully")
	resp := s.respConv.ToTopupWalletResponse(paymentURL, expiresAt)
	return resp, nil
}
