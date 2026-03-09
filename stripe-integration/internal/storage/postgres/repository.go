package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/turtlepavlo/stripe_integration/internal/domain"
	db "github.com/turtlepavlo/stripe_integration/internal/storage/postgres/generated"
)

type Repository struct {
	q *db.Queries
}

func NewRepository(dbtx db.DBTX) *Repository {
	return &Repository{q: db.New(dbtx)}
}

func (r *Repository) AddToWaitlist(ctx context.Context, entry domain.ToWaitlist) error {
	return r.q.CreateWaitlistEntry(ctx, db.CreateWaitlistEntryParams{
		EventID:   entry.EventID,
		UserID:    entry.UserID,
		UserEmail: entry.UserEmail,
	})
}

func (r *Repository) GetNextWaiter(ctx context.Context, eventID int64) (domain.WaitlistEntry, error) {
	row, err := r.q.GetNextWaitlistEntry(ctx, eventID)
	if err != nil {
		return domain.WaitlistEntry{}, err
	}
	return domain.WaitlistEntry{
		ID:        row.ID,
		EventID:   row.EventID,
		UserID:    row.UserID,
		UserEmail: row.UserEmail,
		Status:    row.Status,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

func (r *Repository) GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (domain.Payment, error) {
	row, err := r.q.GetPaymentByOrderID(ctx, orderID)
	if err != nil {
		return domain.Payment{}, err
	}
	return mapPayment(row), nil
}

func (r *Repository) CreatePayment(ctx context.Context, payment domain.Payment) error {
	return r.q.CreatePayment(ctx, db.CreatePaymentParams{
		OrderID:    payment.OrderID,
		UserID:     payment.UserID,
		Amount:     payment.Amount,
		Currency:   payment.Currency,
		ExternalID: pgtype.Text{String: payment.ExternalID, Valid: payment.ExternalID != ""},
		Status:     payment.Status,
	})
}

func (r *Repository) UpdatePaymentStatus(ctx context.Context, payment domain.Payment) (int64, error) {
	tag, err := r.q.UpdatePaymentStatus(ctx, db.UpdatePaymentStatusParams{
		OrderID:    payment.OrderID,
		Status:     payment.Status,
		ExternalID: pgtype.Text{String: payment.ExternalID, Valid: payment.ExternalID != ""},
	})
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *Repository) GetPaymentByExternalID(ctx context.Context, externalID string) (domain.Payment, error) {
	row, err := r.q.GetPaymentByExternalID(ctx, pgtype.Text{String: externalID, Valid: true})
	if err != nil {
		return domain.Payment{}, err
	}
	return mapPayment(row), nil
}

func (r *Repository) UpdatePaymentStatusByExternalID(ctx context.Context, externalID, status string) (int64, error) {
	tag, err := r.q.UpdatePaymentStatusOnlyByExternalID(ctx, db.UpdatePaymentStatusOnlyByExternalIDParams{
		ExternalID: pgtype.Text{String: externalID, Valid: true},
		Status:     status,
	})
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *Repository) MarkAsNotified(ctx context.Context, id int64) error {
	return r.q.UpdateWaitlistStatus(ctx, db.UpdateWaitlistStatusParams{ID: id, Status: "NOTIFIED"})
}

func mapPayment(row db.Payment) domain.Payment {
	externalID := ""
	if row.ExternalID.Valid {
		externalID = row.ExternalID.String
	}

	return domain.Payment{
		ID:         row.ID,
		OrderID:    row.OrderID,
		UserID:     row.UserID,
		Amount:     row.Amount,
		Currency:   row.Currency,
		ExternalID: externalID,
		Status:     row.Status,
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
	}
}
