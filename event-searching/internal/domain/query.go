package domain

import "time"

type SearchFilter struct {
	Query       string
	PerformerID int64
	VenueID     int64
	DateFrom    time.Time
	DateTo      time.Time
	Offset      int64
	Limit       int64
}
