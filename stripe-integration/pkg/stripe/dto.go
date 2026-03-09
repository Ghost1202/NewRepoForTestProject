package stripe

import "time"

type CheckoutSessionParams struct {
	OrderID   string
	UserID    string
	WalletID  int64
	UserEmail string
	Amount    int64
	Currency  string
	TTL       time.Duration
	TicketIDs []int64
}

type CheckoutSessionResult struct {
	URL string
	ID  string
}

type RefundParams struct {
	PaymentIntentID string
	Amount          int64
	UserID          int64
	EventID         int64
	TicketID        int64
	Description     string
}
