package domain

import "time"

type Ticket struct {
	ID         int64
	EventID    int64
	VenueID    int64
	SectorName string
	RowNumber  int64
	SeatNumber int64
	Price      int64
	Currency   string
	Status     string
	UserID     int64
	CreatedAt  time.Time
}
