package service

import (
	"context"

	"github.com/turtlepavlo/event-searching/internal/domain"
	"github.com/turtlepavlo/event-searching/internal/lib/telemetry"
	"go.uber.org/zap"
)

func (s *SearchService) GetRating(ctx context.Context, eventID int64) (domain.Rating, error) {
	const op = "SearchService.GetRating"

	log := telemetry.WithTrace(ctx, s.log).With(
		zap.String("layer", "service"),
		zap.String("op", op),
	)

	row, err := s.commentRepo.GetRatingByEvent(ctx, eventID)
	if err != nil {
		log.Error("get rating failed",
			zap.Int64("event_id", eventID),
			zap.Error(err),
		)
		return domain.Rating{}, err
	}

	return s.toDomain.toDomainRating(row), nil
}
