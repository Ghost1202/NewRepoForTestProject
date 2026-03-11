package server

import (
	"context"

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

func (s *Server) Waitlist(ctx context.Context, req *paymentv1.WaitlistRequest) (*paymentv1.WaitlistResponse, error) {
	const op = "server.Server.Waitlist"

	tracer := otel.Tracer("stripe-integration/internal/transport/grpc/server")
	ctx, span := tracer.Start(ctx, op, trace.WithAttributes(
		attribute.String("rpc.system", "grpc"),
		attribute.String("rpc.service", "payment.v1.PaymentService"),
		attribute.String("rpc.method", "Waitlist"),
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
		attribute.String("user_email", req.GetUserEmail()),
	)

	entry := s.reqConv.ToWaitlistInput(req)

	log.Debug("adding user to waitlist")
	if err := s.service.AddUserToWaitlist(ctx, entry); err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, "AddUserToWaitlist failed")
		log.Error("failed to add user to waitlist", zap.Error(err))
		return nil, status.Error(grpcCodes.Internal, "internal server error")
	}

	log.Info("user added to waitlist")
	return &paymentv1.WaitlistResponse{}, nil
}
