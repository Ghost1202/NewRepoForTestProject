package service

import (
	"context"

	"github.com/turtlepavlo/event-booking/internal/domain"
	"github.com/turtlepavlo/event-booking/pkg/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func (srv *BookingService) GetEventDiscount(ctx context.Context, eventID int64) ([]domain.Early, []domain.Bundle, error) {
	const op = "internal.service.GetEventDiscount"
	tracer := otel.Tracer("/service/booking")

	ctx, span := tracer.Start(ctx, op, trace.WithAttributes(attribute.Int64("event_id", eventID)))
	defer span.End()

	log := telemetry.WithTrace(ctx, srv.log)

	earlyModels, err := srv.tickets.GetEarlyByEvent(ctx, eventID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db error")
		log.Error("failed to fetch early discounts",
			zap.Error(err),
			zap.Int64("event_id", eventID),
		)
		return nil, nil, err
	}

	bundleModels, err := srv.tickets.GetBundleByEvent(ctx, eventID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db error")
		log.Error("failed to fetch bundle discounts",
			zap.Error(err),
			zap.Int64("event_id", eventID),
		)
		return nil, nil, err
	}

	return srv.fromStorage.Earlies(earlyModels), srv.fromStorage.Bundles(bundleModels), nil
}
