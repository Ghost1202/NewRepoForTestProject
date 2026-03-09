package handler

import (
	"context"

	"github.com/turtlepavlo/event-service/internal/domain"
	"go.uber.org/zap"
)

type EventService interface {
	CreateEvent(ctx context.Context, actor domain.Performer, event *domain.Event, sectors []domain.TicketSector, promos []domain.Promo, early []domain.Early, bundles []domain.Bundle) (int64, error)
	UpdateEvent(ctx context.Context, actor domain.Performer, event *domain.Event) error

	AddPromo(ctx context.Context, actor domain.Performer, eventID int64, promos []domain.Promo) ([]int64, error)
	AddEarly(ctx context.Context, actor domain.Performer, eventID int64, early []domain.Early) ([]int64, error)
	AddBundle(ctx context.Context, actor domain.Performer, eventID int64, bundles []domain.Bundle) ([]int64, error)

	GetPromo(ctx context.Context, actor domain.Performer, filter domain.Filter) ([]domain.Promo, error)
	GetEarly(ctx context.Context, actor domain.Performer, filter domain.Filter) ([]domain.Early, error)
	GetBundle(ctx context.Context, actor domain.Performer, filter domain.Filter) ([]domain.Bundle, error)
}

type Handler struct {
	srv      EventService
	log      *zap.Logger
	reqConv  *RequestConverter
	respConv *ResponseConverter
}

func New(srv EventService, log *zap.Logger, reqConv *RequestConverter, respConv *ResponseConverter) *Handler {
	return &Handler{
		srv:      srv,
		reqConv:  reqConv,
		respConv: respConv,
		log:      log,
	}
}
