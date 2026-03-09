package handler

import (
	"context"

	"github.com/turtlepavlo/event-booking/internal/domain"
	transport "github.com/turtlepavlo/event-booking/internal/transport/http"
	"go.uber.org/zap"
)

type BookingService interface {
	GetTickets(ctx context.Context, eventID int64) ([]domain.Ticket, error)
	BookTickets(ctx context.Context, booking domain.Booking) (string, error)
	OrderByWallet(ctx context.Context, booking domain.Booking) (domain.WalletChargeResult, error)
	GetEventDiscount(ctx context.Context, eventID int64) ([]domain.Early, []domain.Bundle, error)

	GetUserTickets(ctx context.Context, userID int64) ([]domain.Ticket, error)
	TransferTicket(ctx context.Context, transfer domain.TicketTransfer) error

	RefundTicket(ctx context.Context, refund domain.Refund) error
	AddToWaitlist(ctx context.Context, item domain.Waitlist) error
}

type Handler struct {
	srv      BookingService
	log      *zap.Logger
	respConv ResponseConverter
	reqConv  RequestConverter
	cfg      transport.Config
}

func New(log *zap.Logger, srv BookingService, respConv ResponseConverter, reqConv RequestConverter, cfg transport.Config) *Handler {
	return &Handler{
		srv:      srv,
		log:      log,
		respConv: respConv,
		reqConv:  reqConv,
		cfg:      cfg,
	}
}
