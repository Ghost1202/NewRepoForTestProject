package domain

import "time"

type Early struct {
	ID         int64
	EventID    int64
	Code       string
	Type       string
	Value      float64
	SectorName string
	ValidUntil time.Time
}
