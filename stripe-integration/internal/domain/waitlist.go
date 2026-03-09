package domain

import "time"

type WaitlistEntry struct {
	ID        int64
	EventID   int64
	UserID    int64
	UserEmail string
	Status    string
	CreatedAt time.Time
}

type ToWaitlist struct {
	UserID    int64
	EventID   int64
	UserEmail string
}
