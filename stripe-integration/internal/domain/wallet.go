package domain

import (
	"time"

	"github.com/google/uuid"
)

type TopupWalletInput struct {
	OrderID  uuid.UUID
	UserID   int64
	WalletID int64
	Amount   int64
	Currency string
	TokenTTL time.Duration
}
