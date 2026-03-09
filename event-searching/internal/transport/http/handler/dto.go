package handler

import "time"

type (
	GetEventsRequest struct {
		Limit  int64 `query:"limit" json:"limit" example:"10"`
		Offset int64 `query:"offset" json:"offset" example:"0"`
	}
	GetEventsResponse struct {
		Events []EventResponse `json:"events"`
		Count  int64           `json:"count" example:"100"`
	}
)

type (
	GetEventsFilterRequest struct {
		Query       string    `query:"query" json:"query" example:"rock concert"`
		PerformerID int64     `query:"performer_id" json:"performer_id" example:"123"`
		VenueID     int64     `query:"venue_id" json:"venue_id" example:"456"`
		DateFrom    time.Time `query:"date_from" json:"date_from" example:"2026-01-01T00:00:00Z"`
		DateTo      time.Time `query:"date_to" json:"date_to" example:"2026-12-31T23:59:59Z"`
		Limit       int64     `query:"limit" json:"limit" example:"20"`
		Offset      int64     `query:"offset" json:"offset" example:"0"`
	}

	GetEventsFilterResponse struct {
		Events []EventResponse `json:"events"`
		Count  int64           `json:"count" example:"5"`
	}
)

type EventResponse struct {
	ID            int64     `json:"id" example:"1"`
	PerformerID   int64     `json:"performer_id" example:"101"`
	VenueID       int64     `json:"venue_id" example:"202"`
	Name          string    `json:"name" example:"Grand Concert"`
	StartDate     time.Time `json:"date_start" example:"2026-05-20T19:00:00Z"`
	SoldOut       bool      `json:"is_sold_out" example:"false"`
	SaleStartDate time.Time `json:"date_sale_start,omitempty" example:"2026-01-01T10:00:00Z"`
	InfoHeader    string    `json:"info_header,omitempty" example:"Important Info"`
	InfoBody      string    `json:"info_body,omitempty" example:"Doors open at 18:00"`
	Popularity    int64     `json:"popularity" example:"150"`
}

type SetCommentRequest struct {
	EventID int64  `json:"event_id" example:"123" minimum:"1"`
	Rating  int64  `json:"rating" example:"48" minimum:"0" maximum:"50"`
	Text    string `json:"text" example:"Great event!"`
	Nick    string `json:"nick" example:"pavlo"`
}

type (
	GetCommentsRequest struct {
		EventID int64 `json:"event_id" example:"123" minimum:"1"`
		Limit   int64 `json:"limit" example:"20" minimum:"1" maximum:"100"`
		Offset  int64 `json:"offset" example:"0" minimum:"0"`
	}

	GetCommentsResponse struct {
		EventID int64     `json:"event_id" example:"123" minimum:"1"`
		Limit   int64     `json:"limit" example:"20" minimum:"1" maximum:"100"`
		Offset  int64     `json:"offset" example:"0" minimum:"0"`
		Items   []Comment `json:"items"`
	}
)

type Comment struct {
	EventID int64  `json:"event_id" example:"123" minimum:"1"`
	UserID  int64  `json:"user_id" example:"777" minimum:"1"`
	Nick    string `json:"nick" example:"pavlo"`
	Text    string `json:"text" example:"Great event!"`
	Rating  int64  `json:"rating" example:"48" minimum:"0" maximum:"50"`
}

type (
	GetRatingRequest struct {
		EventID int64 `json:"event_id" example:"123" minimum:"1"`
	}

	GetRatingResponse struct {
		EventID int64 `json:"event_id" example:"123" minimum:"1"`
		Avg     int64 `json:"avg" example:"43" minimum:"0" maximum:"50"`
		Count   int64 `json:"count" example:"511" minimum:"0"`
	}
)
