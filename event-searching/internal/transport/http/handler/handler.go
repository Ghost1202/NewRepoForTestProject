package handler

import (
	"context"

	"github.com/turtlepavlo/event-searching/internal/domain"
	"go.uber.org/zap"
)

type SearchService interface {
	GetPopularEvents(ctx context.Context, limit, offset int64) ([]domain.Event, error)
	SearchEventsFilter(ctx context.Context, filter domain.SearchFilter) ([]domain.Event, error)

	SetComment(ctx context.Context, comment domain.UpsertComment) error
	GetComments(ctx context.Context, req domain.ListComments) ([]domain.Comment, error)
	GetRating(ctx context.Context, eventID int64) (domain.Rating, error)
}

type Handler struct {
	srv      SearchService
	log      *zap.Logger
	reqConv  *RequestConverter
	respConv *ResponseConverter
}

func New(log *zap.Logger, srv SearchService, reqConv *RequestConverter, respConv *ResponseConverter) *Handler {
	return &Handler{
		srv:      srv,
		log:      log,
		reqConv:  reqConv,
		respConv: respConv,
	}
}
