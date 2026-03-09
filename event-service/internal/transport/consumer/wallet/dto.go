package wallet

import "time"

type ChargeCompletedEvent struct {
	BookingID       string    `json:"booking_id"`
	UserID          int64     `json:"user_id"`
	EventID         int64     `json:"event_id"`
	TicketIDs       []int64   `json:"ticket_ids"`
	Amount          int64     `json:"amount"`
	DiscountPercent float64   `json:"discount_percent"`
	DiscountAmount  int64     `json:"discount_amount"`
	FinalAmount     int64     `json:"final_amount"`
	Currency        string    `json:"currency"`
	OccurredAt      time.Time `json:"occurred_at"`
}
