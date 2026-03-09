package service

import (
	"context"

	"github.com/turtlepavlo/event-service/internal/storage"
	"go.uber.org/zap"
)

const (
	StatusCreated = "CREATED"
	StatusSold    = "SOLD"
	Performer     = "PERFORMER"
)

type EventStore interface {
	CreateEvent(ctx context.Context, event storage.Events, tickets []storage.Tickets, promos []storage.Promo, early []storage.Early, bundles []storage.Bundle) (int64, error)
	UpdateEvent(ctx context.Context, event storage.Events) error
	GetEventPerformerID(ctx context.Context, eventID int64) (int64, error)

	UpdateTicketStatus(ctx context.Context, ticketID int64, fromStatus, toStatus string) (int64, error)
	UpdateTicketOwner(ctx context.Context, ticketID, fromUserID, toUserID int64) (int64, error)
	UpdateTicketStatusAndOwner(ctx context.Context, ticketID int64, status string, userID int64) (int64, error)
	GetTicketMeta(ctx context.Context, ticketID int64) (storage.TicketMeta, bool, error)

	AddPromo(ctx context.Context, promos []storage.Promo) ([]int64, error)
	AddEarly(ctx context.Context, early []storage.Early) ([]int64, error)
	AddBundle(ctx context.Context, bundles []storage.Bundle) ([]int64, error)

	GetPromo(ctx context.Context, eventID, limit, offset int64) ([]storage.Promo, error)
	GetEarly(ctx context.Context, eventID, limit, offset int64) ([]storage.Early, error)
	GetBundle(ctx context.Context, eventID, limit, offset int64) ([]storage.Bundle, error)
}
type EventService struct {
	repo EventStore
	log  *zap.Logger
	conv converte
}

func NewEventService(repo EventStore, log *zap.Logger) *EventService {
	return &EventService{
		repo: repo,
		log:  log,
		conv: converte{},
	}
}
