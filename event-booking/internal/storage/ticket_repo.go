package storage

import (
	"context"

	"github.com/turtlepavlo/event-booking/pkg/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const StatusCreated = "CREATED"

type TicketRepo struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewTicketRepo(db *gorm.DB, log *zap.Logger) *TicketRepo {
	return &TicketRepo{
		db:  db,
		log: log,
	}
}

func (repo *TicketRepo) GetTicketsByEvent(ctx context.Context, eventID int64) ([]TicketModel, error) {
	const op = "TicketRepo.GetCreatedByEvent"
	tracer := otel.Tracer("storage/ticket_repo")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
			attribute.Int64("event_id", eventID),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log)

	var models []TicketModel

	err := repo.db.WithContext(ctx).
		Where("event_id = ? AND status = ?", eventID, StatusCreated).
		Find(&models).Error
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db select failed")
		log.Error("postgres fetch failed",
			zap.Error(err),
			zap.Int64("event_id", eventID),
		)
		return nil, err
	}

	if len(models) == 0 {
		return []TicketModel{}, nil
	}

	return models, nil
}

func (repo *TicketRepo) GetTicketByID(ctx context.Context, ticketID int64) (TicketModel, error) {
	const op = "TicketRepo.GetTicketByID"
	tracer := otel.Tracer("storage/ticket_repo")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
			attribute.Int64("ticket_id", ticketID),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log)

	var model TicketModel
	err := repo.db.WithContext(ctx).First(&model, ticketID).Error
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "record not found or db error")
		log.Error("failed to get ticket",
			zap.Error(err),
			zap.Int64("ticket_id", ticketID),
		)
		return TicketModel{}, err
	}

	return model, nil
}

func (repo *TicketRepo) GetTicketsByUserID(ctx context.Context, userID int64) ([]TicketModel, error) {
	const op = "TicketRepo.GetTicketsByUserID"
	tracer := otel.Tracer("storage/ticket_repo")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
			attribute.Int64("user_id", userID),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log)

	var models []TicketModel
	err := repo.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db select failed")
		log.Error("postgres fetch failed",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		return nil, err
	}

	if len(models) == 0 {
		return []TicketModel{}, nil
	}

	return models, nil
}

func (repo *TicketRepo) GetPromoByCode(ctx context.Context, eventID int64, code string) (Promo, error) {
	const op = "TicketRepo.GetPromoByCode"
	tracer := otel.Tracer("storage/ticket_repo")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
			attribute.Int64("event_id", eventID),
			attribute.String("promo.code", code),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log)

	var model Promo
	err := repo.db.WithContext(ctx).
		Where("event_id = ? AND code = ?", eventID, code).
		First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Debug("promo code not found",
				zap.Int64("event_id", eventID),
				zap.String("code", code),
			)
			return Promo{}, err
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "db select failed")
		log.Error("failed to fetch promo code",
			zap.Error(err),
			zap.Int64("event_id", eventID),
			zap.String("code", code),
		)
		return Promo{}, err
	}

	return model, nil
}

func (repo *TicketRepo) GetTicketsByIDs(ctx context.Context, ticketIDs []int64) ([]TicketModel, error) {
	const op = "TicketRepo.GetTicketsByIDs"
	tracer := otel.Tracer("storage/ticket_repo")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
			attribute.Int64("tickets.count", int64(len(ticketIDs))),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log)

	var models []TicketModel
	err := repo.db.WithContext(ctx).
		Where("id IN ?", ticketIDs).
		Find(&models).Error
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db select failed")
		log.Error("failed to fetch tickets by ids", zap.Error(err))
		return nil, err
	}

	if len(models) == 0 {
		return []TicketModel{}, nil
	}

	return models, nil
}

func (repo *TicketRepo) GetEarlyByEvent(ctx context.Context, eventID int64) ([]Early, error) {
	const op = "TicketRepo.GetEarlyByEvent"
	tracer := otel.Tracer("storage/ticket_repo")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
			attribute.Int64("event_id", eventID),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log)

	var models []Early
	err := repo.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		Find(&models).Error
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db select failed")
		log.Error("failed to fetch early offers",
			zap.Error(err),
			zap.Int64("event_id", eventID),
		)
		return nil, err
	}

	if len(models) == 0 {
		return []Early{}, nil
	}

	return models, nil
}

func (repo *TicketRepo) GetBundleByEvent(ctx context.Context, eventID int64) ([]Bundle, error) {
	const op = "TicketRepo.GetBundleByEvent"
	tracer := otel.Tracer("storage/ticket_repo")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
			attribute.Int64("event_id", eventID),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log)

	var models []Bundle
	err := repo.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		Find(&models).Error
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db select failed")
		log.Error("failed to fetch bundle offers",
			zap.Error(err),
			zap.Int64("event_id", eventID),
		)
		return nil, err
	}

	if len(models) == 0 {
		return []Bundle{}, nil
	}

	return models, nil
}

func (repo *TicketRepo) GetEarlyByID(ctx context.Context, earlyID int64) (Early, error) {
	const op = "TicketRepo.GetEarlyByID"
	tracer := otel.Tracer("storage/ticket_repo")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
			attribute.Int64("early_id", earlyID),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log)

	var model Early
	err := repo.db.WithContext(ctx).First(&model, earlyID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Debug("early offer not found",
				zap.Int64("early_id", earlyID),
			)
			return Early{}, err
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "db select failed")
		log.Error("failed to fetch early offer",
			zap.Error(err),
			zap.Int64("early_id", earlyID),
		)
		return Early{}, err
	}

	return model, nil
}

func (repo *TicketRepo) GetBundleByID(ctx context.Context, bundleID int64) (Bundle, error) {
	const op = "TicketRepo.GetBundleByID"
	tracer := otel.Tracer("storage/ticket_repo")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
			attribute.Int64("bundle_id", bundleID),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log)

	var model Bundle
	err := repo.db.WithContext(ctx).First(&model, bundleID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Debug("bundle offer not found",
				zap.Int64("bundle_id", bundleID),
			)
			return Bundle{}, err
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "db select failed")
		log.Error("failed to fetch bundle offer",
			zap.Error(err),
			zap.Int64("bundle_id", bundleID),
		)
		return Bundle{}, err
	}

	return model, nil
}
