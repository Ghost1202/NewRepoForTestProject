package service

import (
	"context"

	"github.com/turtlepavlo/event-searching/internal/storage"
	sqlc "github.com/turtlepavlo/event-searching/internal/storage/postgres/sqlc"
	"go.uber.org/zap"
)

type FavouriteStore interface {
	GetPopular(ctx context.Context, limit, offset int64) ([]storage.EventModel, error)
	PromoteToPopular(ctx context.Context, events []storage.EventModel) error
}

type SearchRepo interface {
	Search(ctx context.Context, filter storage.SearchFilterDTO) ([]storage.EventModel, error)
}

type CommentRepo interface {
	CreateComment(ctx context.Context, comment sqlc.CreateCommentParams) (sqlc.CreateCommentRow, error)
	ListCommentsByEvent(ctx context.Context, req sqlc.ListCommentsByEventParams) ([]sqlc.EventComment, error)
	GetRatingByEvent(ctx context.Context, eventID int64) (sqlc.GetRatingByEventRow, error)
}

type SearchService struct {
	cfg           Config
	favouriteRepo FavouriteStore
	eventRepo     SearchRepo
	commentRepo   CommentRepo
	log           *zap.Logger
	toDomain      converteToDomain
	toStore       converterToStorage
}

func NewSearchService(cfg Config, favouriteRepo FavouriteStore, eventRepo SearchRepo, commentRepo CommentRepo, log *zap.Logger) *SearchService {
	return &SearchService{
		cfg:           cfg,
		favouriteRepo: favouriteRepo,
		eventRepo:     eventRepo,
		commentRepo:   commentRepo,
		log:           log,
		toDomain:      converteToDomain{},
		toStore:       converterToStorage{},
	}
}
