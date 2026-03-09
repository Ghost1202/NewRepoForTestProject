package server

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"
	paymentv1 "github.com/turtlepavlo/proto-contract/gen/go/payment/v1"
	"github.com/turtlepavlo/stripe_integration/internal/domain"
)

type PaymentService interface {
	CreatePaymentLink(ctx context.Context, input domain.CreatePaymentInput) (string, time.Time, error)
	DepositWallet(ctx context.Context, input domain.TopupWalletInput) (string, time.Time, error)
	ConfirmPayment(ctx context.Context, orderID uuid.UUID, sessionID string) error

	CreateRefund(ctx context.Context, refund domain.Refund) error
	AddUserToWaitlist(ctx context.Context, entry domain.ToWaitlist) error
}

type Server struct {
	paymentv1.UnimplementedPaymentServiceServer

	service  PaymentService
	log      *zap.Logger
	reqConv  *RequestConverter
	respConv *ResponseConverter
}

func New(service PaymentService, log *zap.Logger) *Server {
	return &Server{
		service: service,
		log: log.With(
			zap.String("layer", "transport"),
			zap.String("component", "grpc_server"),
		),
		reqConv:  NewRequestConverter(),
		respConv: NewResponseConverter(),
	}
}
