package domain

import "time"

// NOTE ABOUT RATING FORMAT
// We store and transport `Rating` / `RatingAvg` as an integer in "tenths" (0..50), not as a float.
// Example: 48 == 4.8, 6 == 0.6, 40 == 4.0.
// Client must divide by 10 to display the human-readable value (0.0..5.0 with one decimal place).
//
// Rationale:
// - Avoids floating-point precision issues (e.g., 4.8 becoming 4.7999999).
// - Keeps aggregates exact and stable for large volumes (sum/count stays integer).
// - Makes validation trivial (integer range check 0..50).

type Comment struct {
	ID        int64
	EventID   int64
	UserID    int64
	Nick      string
	Text      string
	Rating    int64
	CreatedAt time.Time
}

type UpsertComment struct {
	EventID int64
	UserID  int64
	Nick    string
	Text    string
	Rating  int64
}

type ListComments struct {
	EventID int64
	Limit   int64
	Offset  int64
}

type Rating struct {
	EventID int64
	Avg     int64
	Count   int64
}
