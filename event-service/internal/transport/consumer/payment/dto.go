package payment

import "time"

type ConfirmedEvent struct {
	OrderID    string    `json:"order_id"`
	Amount     int64     `json:"amount"`
	Currency   string    `json:"currency"`
	PaymentID  string    `json:"payment_id"`
	Status     string    `json:"status"`
	OccurredAt time.Time `json:"occurred_at"`
}

type RefundedEvent struct {
	UserID     int64     `json:"user_id"`
	EventID    int64     `json:"event_id"`
	TicketID   int64     `json:"ticket_id"`
	Amount     int64     `json:"amount"`
	Currency   string    `json:"currency"`
	Status     string    `json:"status"`
	OccurredAt time.Time `json:"occurred_at"`
}
