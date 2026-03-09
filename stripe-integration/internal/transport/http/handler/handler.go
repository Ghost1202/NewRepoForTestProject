package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/turtlepavlo/stripe_integration/internal/domain"
	cfg "github.com/turtlepavlo/stripe_integration/internal/transport/http"
	"go.uber.org/zap"
)

const (
	CheckoutSessionCompleted = "checkout.session.completed"
	RefundUpdated            = "refund.updated"
)

type PaymentService interface {
	ConfirmPayment(ctx context.Context, orderID uuid.UUID, sessionID string) error
	ConfirmDeposit(ctx context.Context, orderID uuid.UUID, sessionID string) error
	ConfirmRefund(ctx context.Context, externalID string, refund domain.Refund) error
}

type Handler struct {
	srv  PaymentService
	log  *zap.Logger
	cfg  cfg.Config
	toDM convertToDomain
}

func NewHandler(log *zap.Logger, srv PaymentService, cfg cfg.Config) *Handler {
	return &Handler{
		srv:  srv,
		log:  log,
		cfg:  cfg,
		toDM: newConvertToDomain(),
	}
}
