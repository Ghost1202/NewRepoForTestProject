package storage

import (
	"context"

	"github.com/turtlepavlo/event-service/internal/lib/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Repository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewRepository(db *gorm.DB, log *zap.Logger) *Repository {
	return &Repository{
		db:  db,
		log: log,
	}
}

func (repo *Repository) CreateEvent(ctx context.Context, event Events, tickets []Tickets, promos []Promo, early []Early, bundles []Bundle) (int64, error) {
	const op = "Repository.CreateEventWithTicketsAndPromos"
	tracer := otel.Tracer("internal/repository")

	promoCount := len(promos) + len(early) + len(bundles)
	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "TX_INSERT_FULL_EVENT"),
			attribute.String("event_name", event.Name),
			attribute.Int("tickets_count", len(tickets)),
			attribute.Int("promos_count", promoCount),
		),
	)
	defer span.End()

	var eventID int64

	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(&event).Error; err != nil {
			return err
		}
		eventID = event.ID
		for i := range tickets {
			tickets[i].EventID = eventID
		}

		if len(tickets) > 0 {
			if err := tx.WithContext(ctx).CreateInBatches(&tickets, 1000).Error; err != nil {
				return err
			}
		}
		if len(promos) > 0 {
			for i := range promos {
				promos[i].EventID = eventID
			}
			if err := tx.WithContext(ctx).CreateInBatches(&promos, 100).Error; err != nil {
				return err
			}
		}
		if len(early) > 0 {
			for i := range early {
				early[i].EventID = eventID
			}
			if err := tx.WithContext(ctx).CreateInBatches(&early, 100).Error; err != nil {
				return err
			}
		}
		if len(bundles) > 0 {
			for i := range bundles {
				bundles[i].EventID = eventID
			}
			if err := tx.WithContext(ctx).CreateInBatches(&bundles, 100).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "transaction failed")
		telemetry.WithTrace(ctx, repo.log).Error("failed to create event with tickets and promos",
			zap.String("event_name", event.Name),
			zap.Error(err),
		)
		return 0, err
	}

	span.SetAttributes(attribute.Int64("event_id", eventID))
	telemetry.WithTrace(ctx, repo.log).Info("event created successfully",
		zap.Int64("event_id", eventID),
		zap.Int("tickets_count", len(tickets)),
		zap.Int("promos_count", promoCount),
	)

	return eventID, nil
}

func (repo *Repository) UpdateEvent(ctx context.Context, event Events) error {
	const op = "Repository.UpdateEvent"
	tracer := otel.Tracer("internal/repository")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "UPDATE"),
			attribute.Int64("event_id", event.ID),
		),
	)
	defer span.End()

	if err := repo.db.WithContext(ctx).
		Model(&Events{ID: event.ID}).
		Updates(&event).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "update failed")
		telemetry.WithTrace(ctx, repo.log).Error("failed to update event",
			zap.Int64("event_id", event.ID),
			zap.Error(err),
		)
		return err
	}

	telemetry.WithTrace(ctx, repo.log).Debug("event updated successfully")
	return nil
}

func (repo *Repository) UpdateTicketStatus(ctx context.Context, ticketID int64, fromStatus, toStatus string) (int64, error) {
	const op = "Repository.UpdateTicketStatus"
	tracer := otel.Tracer("internal/repository")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "UPDATE"),
			attribute.Int64("ticket_id", ticketID),
		),
	)
	defer span.End()

	result := repo.db.WithContext(ctx).
		Model(&Tickets{}).
		Where("id = ? AND status = ?", ticketID, fromStatus).
		Update("status", toStatus)

	if result.Error != nil {
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, "update failed")
		telemetry.WithTrace(ctx, repo.log).Error("failed to update ticket status", zap.Error(result.Error))
		return 0, result.Error
	}

	span.SetAttributes(attribute.Int64("db.rows_affected", result.RowsAffected))
	return result.RowsAffected, nil
}

func (repo *Repository) GetTicketMeta(ctx context.Context, ticketID int64) (TicketMeta, bool, error) {
	const op = "Repository.GetTicketMeta"
	tracer := otel.Tracer("internal/repository")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
			attribute.Int64("ticket_id", ticketID),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log).With(
		zap.String("op", op),
		zap.Int64("ticket_id", ticketID),
	)

	var meta TicketMeta
	tx := repo.db.WithContext(ctx).
		Model(&Tickets{}).
		Select(`"id", "event_id", "status", "user_id"`).
		Where("id = ?", ticketID).
		Scan(&meta)

	if tx.Error != nil {
		span.RecordError(tx.Error)
		span.SetStatus(codes.Error, "select failed")
		log.Error("failed to get ticket meta", zap.Error(tx.Error))
		return TicketMeta{}, false, tx.Error
	}
	if tx.RowsAffected == 0 {
		log.Debug("ticket meta not found (empty result)")
		return TicketMeta{}, false, nil
	}

	span.SetAttributes(
		attribute.Int64("event_id", meta.EventID),
		attribute.String("status", meta.Status),
		attribute.Int64("user_id", meta.UserID),
	)
	log.Debug("ticket meta fetched",
		zap.Int64("event_id", meta.EventID),
		zap.String("status", meta.Status),
		zap.Int64("user_id", meta.UserID),
	)

	return meta, true, nil
}

func (repo *Repository) UpdateTicketOwner(ctx context.Context, ticketID, fromUserID, toUserID int64) (int64, error) {
	const op = "Repository.UpdateTicketOwner"
	tracer := otel.Tracer("internal/repository")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "UPDATE"),
			attribute.Int64("ticket_id", ticketID),
			attribute.Int64("from_user_id", fromUserID),
			attribute.Int64("to_user_id", toUserID),
		),
	)
	defer span.End()

	result := repo.db.WithContext(ctx).
		Model(&Tickets{}).
		Where("id = ? AND user_id = ?", ticketID, fromUserID).
		Update("user_id", toUserID)
	if result.Error != nil {
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, "update failed")
		telemetry.WithTrace(ctx, repo.log).Error("failed to update ticket owner", zap.Error(result.Error))
		return 0, result.Error
	}

	span.SetAttributes(attribute.Int64("db.rows_affected", result.RowsAffected))
	return result.RowsAffected, nil
}

func (repo *Repository) UpdateTicketStatusAndOwner(ctx context.Context, ticketID int64, status string, userID int64) (int64, error) {
	const op = "Repository.UpdateTicketStatusAndOwner"
	tracer := otel.Tracer("internal/repository")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "UPDATE"),
			attribute.Int64("ticket_id", ticketID),
		),
	)
	defer span.End()

	result := repo.db.WithContext(ctx).
		Model(&Tickets{}).
		Where("id = ?", ticketID).
		Updates(map[string]interface{}{
			"status":  status,
			"user_id": userID,
		})

	if result.Error != nil {
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, "update failed")
		telemetry.WithTrace(ctx, repo.log).Error("failed to update ticket", zap.Error(result.Error))
		return 0, result.Error
	}

	span.SetAttributes(attribute.Int64("db.rows_affected", result.RowsAffected))
	return result.RowsAffected, nil
}

func (repo *Repository) GetEventPerformerID(ctx context.Context, eventID int64) (int64, error) {
	const op = "Repository.GetEventPerformerID"
	tracer := otel.Tracer("internal/repository")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
			attribute.Int64("event_id", eventID),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log).With(
		zap.String("op", op),
		zap.Int64("event_id", eventID),
	)

	var rrow row

	tx := repo.db.WithContext(ctx).
		Model(&Events{}).
		Select(`"performer_id"`).
		Where("id = ?", eventID).
		Scan(&rrow)

	if tx.Error != nil {
		span.RecordError(tx.Error)
		span.SetStatus(codes.Error, "select failed")
		log.Error("failed to get event performer id", zap.Error(tx.Error))
		return 0, tx.Error
	}

	if tx.RowsAffected == 0 {
		log.Debug("event performer id not found (empty result)")
		return 0, nil
	}

	span.SetAttributes(attribute.Int64("performer_id", rrow.PerformerID))
	log.Debug("event performer id fetched", zap.Int64("performer_id", rrow.PerformerID))

	return rrow.PerformerID, nil
}

func (repo *Repository) AddPromo(ctx context.Context, promos []Promo) ([]int64, error) {
	const op = "Repository.AddPromo"
	tracer := otel.Tracer("internal/repository")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "INSERT_BATCH"),
			attribute.Int("count", len(promos)),
		),
	)
	defer span.End()

	if len(promos) == 0 {
		return nil, nil
	}

	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tx = tx.Session(&gorm.Session{SkipDefaultTransaction: true})
		if err := tx.WithContext(ctx).CreateInBatches(&promos, 100).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "insert failed")
		telemetry.WithTrace(ctx, repo.log).Error("failed to add promos", zap.Error(err))
		return nil, err
	}

	promoIDs := make([]int64, len(promos))
	for i := range promos {
		promoIDs[i] = promos[i].ID
	}

	telemetry.WithTrace(ctx, repo.log).Info("promos added successfully", zap.Int("count", len(promos)))
	return promoIDs, nil
}

func (repo *Repository) AddEarly(ctx context.Context, early []Early) ([]int64, error) {
	const op = "Repository.AddEarly"
	tracer := otel.Tracer("internal/repository")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "INSERT_BATCH"),
			attribute.Int("count", len(early)),
		),
	)
	defer span.End()

	if len(early) == 0 {
		return nil, nil
	}

	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tx = tx.Session(&gorm.Session{SkipDefaultTransaction: true})
		if err := tx.WithContext(ctx).CreateInBatches(&early, 100).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "insert failed")
		telemetry.WithTrace(ctx, repo.log).Error("failed to add early promos", zap.Error(err))
		return nil, err
	}

	earlyIDs := make([]int64, len(early))
	for i := range early {
		earlyIDs[i] = early[i].ID
	}

	telemetry.WithTrace(ctx, repo.log).Info("early promos added successfully", zap.Int("count", len(early)))
	return earlyIDs, nil
}

func (repo *Repository) AddBundle(ctx context.Context, bundles []Bundle) ([]int64, error) {
	const op = "Repository.AddBundle"
	tracer := otel.Tracer("internal/repository")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "INSERT_BATCH"),
			attribute.Int("count", len(bundles)),
		),
	)
	defer span.End()

	if len(bundles) == 0 {
		return nil, nil
	}

	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tx = tx.Session(&gorm.Session{SkipDefaultTransaction: true})
		if err := tx.WithContext(ctx).CreateInBatches(&bundles, 100).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "insert failed")
		telemetry.WithTrace(ctx, repo.log).Error("failed to add bundle promos", zap.Error(err))
		return nil, err
	}

	bundleIDs := make([]int64, len(bundles))
	for i := range bundles {
		bundleIDs[i] = bundles[i].ID
	}

	telemetry.WithTrace(ctx, repo.log).Info("bundle promos added successfully", zap.Int("count", len(bundles)))
	return bundleIDs, nil
}

func (repo *Repository) GetPromo(ctx context.Context, eventID, limit, offset int64) ([]Promo, error) {
	const op = "Repository.GetPromo"
	tracer := otel.Tracer("internal/repository")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
			attribute.Int64("filter_event_id", eventID),
		),
	)
	defer span.End()

	var promos []Promo

	query := repo.db.WithContext(ctx).Model(&Promo{})
	if eventID != 0 {
		query = query.Where("event_id = ?", eventID)
	}

	err := query.
		Limit(int(limit)).
		Offset(int(offset)).
		Order("id DESC").
		Find(&promos).Error
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "select failed")
		telemetry.WithTrace(ctx, repo.log).Error("failed to get promo codes", zap.Error(err))
		return nil, err
	}

	return promos, nil
}

func (repo *Repository) GetEarly(ctx context.Context, eventID, limit, offset int64) ([]Early, error) {
	const op = "Repository.GetEarly"
	tracer := otel.Tracer("internal/repository")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
			attribute.Int64("filter_event_id", eventID),
		),
	)
	defer span.End()

	var early []Early

	query := repo.db.WithContext(ctx).Model(&Early{})
	if eventID != 0 {
		query = query.Where("event_id = ?", eventID)
	}

	err := query.
		Limit(int(limit)).
		Offset(int(offset)).
		Order("id DESC").
		Find(&early).Error
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "select failed")
		telemetry.WithTrace(ctx, repo.log).Error("failed to get early promos", zap.Error(err))
		return nil, err
	}

	return early, nil
}

func (repo *Repository) GetBundle(ctx context.Context, eventID, limit, offset int64) ([]Bundle, error) {
	const op = "Repository.GetBundle"
	tracer := otel.Tracer("internal/repository")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
			attribute.Int64("filter_event_id", eventID),
		),
	)
	defer span.End()

	var bundles []Bundle

	query := repo.db.WithContext(ctx).Model(&Bundle{})
	if eventID != 0 {
		query = query.Where("event_id = ?", eventID)
	}

	err := query.
		Limit(int(limit)).
		Offset(int(offset)).
		Order("id DESC").
		Find(&bundles).Error
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "select failed")
		telemetry.WithTrace(ctx, repo.log).Error("failed to get bundle promos", zap.Error(err))
		return nil, err
	}

	return bundles, nil
}
