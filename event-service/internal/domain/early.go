package domain

import "time"

type Early struct {
	ID         int64
	EventID    int64
	Sector     string
	Code       string
	Type       string
	Value      float64
	ValidUntil time.Time
}
