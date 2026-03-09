package storage

import "time"

type EventModel struct {
	ID            int64     `json:"id"`
	PerformerID   int64     `json:"performer_id"`
	VenueID       int64     `json:"venue_id"`
	Name          string    `json:"name"`
	StartDate     time.Time `json:"date_start"`
	SoldOut       bool      `json:"is_sold_out"`
	SaleStartDate time.Time `json:"date_sale_start,omitempty"`
	InfoHeader    string    `json:"info_header,omitempty"`
	InfoBody      string    `json:"info_body,omitempty"`
	Popularity    int64     `json:"popularity,omitempty"`
}

type PartialModel struct {
	ID            int64     `json:"id"`
	PerformerID   int64     `json:"performer_id,omitempty"`
	VenueID       int64     `json:"venue_id,omitempty"`
	Name          string    `json:"name,omitempty"`
	StartDate     time.Time `json:"date_start,omitempty"`
	SoldOut       bool      `json:"is_sold_out,omitempty"`
	SaleStartDate time.Time `json:"date_sale_start,omitempty"`
	InfoHeader    string    `json:"info_header,omitempty"`
	InfoBody      string    `json:"info_body,omitempty"`
	Popularity    int64     `json:"popularity,omitempty"`
}

type SearchFilterDTO struct {
	Query       string
	PerformerID int64
	VenueID     int64
	DateFrom    time.Time
	DateTo      time.Time
	Limit       int64
	Offset      int64
}
