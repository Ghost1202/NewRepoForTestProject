package service

import (
	"context"
	"errors"

	"github.com/turtlepavlo/event-searching/internal/domain"
)

var ErrInvalidPagination = errors.New("invalid pagination")

func (s *SearchService) GetComments(ctx context.Context, req domain.ListComments) ([]domain.Comment, error) {
	limit := req.Limit
	offset := req.Offset

	if limit == 0 {
		limit = s.cfg.CommentsDefaultLimit
	}
	if limit < 0 || limit > s.cfg.CommentsMaxLimit {
		return nil, ErrInvalidPagination
	}
	if offset < 0 || offset > s.cfg.CommentsMaxOffset {
		return nil, ErrInvalidPagination
	}

	params := s.toStore.toListCommentsParamsFromParts(req.EventID, limit, offset)
	items, err := s.commentRepo.ListCommentsByEvent(ctx, params)
	if err != nil {
		return nil, err
	}

	return s.toDomain.toDomainComments(items), nil
}
