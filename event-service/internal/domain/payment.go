package domain

import "time"

type PaymentEvent struct {
	BookingID string
	Status    string
	Amount    int64
	Timestamp time.Time
}
