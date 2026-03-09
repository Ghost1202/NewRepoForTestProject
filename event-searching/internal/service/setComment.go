package service

import (
	"context"

	"github.com/turtlepavlo/event-searching/internal/domain"
	"github.com/turtlepavlo/event-searching/internal/lib/telemetry"
	"go.uber.org/zap"
)

func (s *SearchService) SetComment(ctx context.Context, comment domain.UpsertComment) error {
	const op = "SearchService.SetComment"
	log := telemetry.WithTrace(ctx, s.log).With(
		zap.String("layer", "service"),
		zap.String("op", op),
	)

	params := s.toStore.toCreateCommentParams(comment)
	_, err := s.commentRepo.CreateComment(ctx, params)
	if err != nil {
		log.Error("create comment failed",
			zap.Int64("event_id", comment.EventID),
			zap.Int64("user_id", comment.UserID),
			zap.Error(err),
		)
		return err
	}

	return nil
}
