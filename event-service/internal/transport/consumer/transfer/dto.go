package transfer

import "time"

type TicketTransfer struct {
	TicketID   int64     `json:"ticket_id"`
	FromUserID int64     `json:"from_user_id"`
	ToUserID   int64     `json:"to_user_id"`
	OccurredAt time.Time `json:"occurred_at"`
}
