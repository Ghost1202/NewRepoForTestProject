package domain

import (
	"time"
)

type Event struct {
	ID            int64
	PerformerID   int64
	VenueID       int64
	Name          string
	StartDate     time.Time
	PostDate      time.Time
	SaleStartDate time.Time
	SoldOut       bool
	MaxPriceCof   float64
	MinPriceCof   float64
}
