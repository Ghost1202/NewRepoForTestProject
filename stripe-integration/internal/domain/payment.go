package domain

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID         int64
	OrderID    uuid.UUID // order UUID from booking service
	UserID     int64
	Amount     int64  // payment amount in cents
	Currency   string // ISO currency code, "USD"
	ExternalID string // transaction ID from payment
	Status     string // PENDING, PAID, FAILED,
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CreatePaymentInput struct {
	OrderID     uuid.UUID
	UserID      int64
	UserEmail   string
	Amount      int64
	Currency    string
	Description string
	TokenTTL    time.Duration
	TicketIDs   []int64
}
