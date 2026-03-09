package domain

import "time"

type Event struct {
	ID            int64
	PerformerID   int64
	VenueID       int64
	Name          string
	StartDate     time.Time
	SoldOut       bool
	SaleStartDate time.Time
	InfoHeader    string
	InfoBody      string
	Popularity    int64
}
