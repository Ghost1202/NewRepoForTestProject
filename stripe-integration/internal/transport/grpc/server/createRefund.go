package server

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	grpcCodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	paymentv1 "github.com/turtlepavlo/proto-contract/gen/go/payment/v1"
	"github.com/turtlepavlo/stripe_integration/pkg/telemetry"
)

func (s *Server) CreateRefund(ctx context.Context, req *paymentv1.CreateRefundRequest) (*paymentv1.CreateRefundResponse, error) {
	const op = "server.Server.CreateRefund"

	tracer := otel.Tracer("/grpc/server")
	ctx, span := tracer.Start(ctx, op, trace.WithAttributes(
		attribute.String("rpc.system", "grpc"),
		attribute.String("rpc.service", "payment.v1.PaymentService"),
		attribute.String("rpc.method", "CreateRefund"),
	))
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
		attribute.Int64("user_id", req.GetUserId()),
		attribute.Int64("event_id", req.GetEventId()),
		attribute.Int64("ticket_id", req.GetTicketId()),
		attribute.Int64("amount", req.GetAmount()),
	)

	refund := s.reqConv.ToCreateRefundInput(req)

	log.Debug("creating refund")
	err := s.service.CreateRefund(ctx, refund)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, "CreateRefundLink failed")

		switch {
		case errors.Is(err, ErrRefundAlreadyProcessed):
			log.Warn("refund already processed", zap.Error(err))
			return nil, status.Error(grpcCodes.FailedPrecondition, "refund already processed")
		case errors.Is(err, ErrRefundLinkExists):
			log.Warn("refund link already exists", zap.Error(err))
			return nil, status.Error(grpcCodes.AlreadyExists, "refund link already exists")
		default:
			log.Error("internal error while creating refund", zap.Error(err))
			return nil, status.Error(grpcCodes.Internal, "internal server error")
		}
	}

	log.Info("refund created successfully")
	return &paymentv1.CreateRefundResponse{}, nil
}
