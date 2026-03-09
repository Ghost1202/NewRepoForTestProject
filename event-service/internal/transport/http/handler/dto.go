package handler

import "time"

type (
	CreateEventReq struct {
		VenueID       int64     `json:"venue_id" example:"33"`
		Name          string    `json:"name" example:"Grand Rock Concert"`
		DateStart     time.Time `json:"date_start" example:"2026-10-05T19:30:00Z"`
		PostDate      time.Time `json:"post_date" example:"2026-09-01T10:00:00Z"`
		SaleStartDate time.Time `json:"sale_start_date" example:"2026-09-05T09:00:00Z"`
		MaxPriceCof   float64   `json:"max_price_cof" example:"2.5"`
		MinPriceCof   float64   `json:"min_price_cof" example:"1.0"`

		Constructor []TicketConstructor `json:"constructor,omitempty"`
		Promos      []Promo             `json:"promos,omitempty"`
		Early       []Early             `json:"early,omitempty"`
		Bundles     []Bundle            `json:"bundles,omitempty"`
	}

	CreateEventResp struct {
		EventID int64 `json:"event_id" example:"5501"`
	}
)

type TicketConstructor struct {
	Name        string `json:"name" example:"VIP"`
	Type        string `json:"type" example:"vip"`
	Price       int64  `json:"price" example:"5000"`
	RowsCount   int64  `json:"rows_count" example:"3"`
	SeatsPerRow int64  `json:"seats_per_row" example:"12"`
}

type UpdateEventReq struct {
	EventID     int64     `json:"event_id" example:"5501"`
	Name        string    `json:"name" example:"Grand Rock Concert - Rescheduled"`
	DateStart   time.Time `json:"date_start" example:"2026-10-06T19:30:00Z"`
	MaxPriceCof float64   `json:"max_price_cof" example:"3.0"`
	MinPriceCof float64   `json:"min_price_cof" example:"1.2"`
}

type Promo struct {
	Code   string  `json:"code" example:"PROMO10"`
	Type   string  `json:"type" example:"percent"`
	Value  float64 `json:"value" example:"10.0"`
	Sector string  `json:"sector,omitempty" example:"VIP"`
}

type Early struct {
	Code       string    `json:"code" example:"EARLYBIRD"`
	Type       string    `json:"type" example:"percent"`
	Value      float64   `json:"value" example:"15.0"`
	Sector     string    `json:"sector,omitempty" example:"VIP"`
	ValidUntil time.Time `json:"valid_until" example:"2026-09-10T09:00:00Z"`
}

type Bundle struct {
	Code     string `json:"code" example:"B2G1"`
	Sector   string `json:"sector,omitempty" example:"VIP"`
	BuyCount int64  `json:"buy_count,omitempty" example:"2"`
	GetCount int64  `json:"get_count,omitempty" example:"1"`
}

type (
	GetPromoReq struct {
		EventID int64 `json:"event_id,omitempty" example:"5501"`
		Limit   int64 `json:"limit" example:"10"`
		Offset  int64 `json:"offset" example:"0"`
	}

	GetPromoRes struct {
		Promos []Promo `json:"promos"`
	}
)

type (
	GetEarlyReq struct {
		EventID int64 `json:"event_id,omitempty" example:"5501"`
		Limit   int64 `json:"limit" example:"10"`
		Offset  int64 `json:"offset" example:"0"`
	}

	GetEarlyRes struct {
		Early []Early `json:"early"`
	}
)

type (
	GetBundleReq struct {
		EventID int64 `json:"event_id,omitempty" example:"5501"`
		Limit   int64 `json:"limit" example:"10"`
		Offset  int64 `json:"offset" example:"0"`
	}

	GetBundleRes struct {
		Bundles []Bundle `json:"bundles"`
	}
)

type (
	AddPromoReq struct {
		EventID int64   `json:"event_id" example:"5501"`
		Promos  []Promo `json:"promos,omitempty"`
	}

	AddPromoResp struct {
		PromoIDs []int64 `json:"promo_ids,omitempty" example:"88"`
	}
)

type (
	AddEarlyReq struct {
		EventID int64   `json:"event_id" example:"5501"`
		Early   []Early `json:"early,omitempty"`
	}

	AddEarlyResp struct {
		EarlyIDs []int64 `json:"early_ids,omitempty" example:"12"`
	}
)

type (
	AddBundleReq struct {
		EventID int64    `json:"event_id" example:"5501"`
		Bundles []Bundle `json:"bundles,omitempty"`
	}

	AddBundleResp struct {
		BundleIDs []int64 `json:"bundle_ids,omitempty" example:"7"`
	}
)
