package handler

import "time"

type (
	CreateBookingRequest struct {
		TicketIDs []int64 `json:"ticket_ids" example:"[5501,5502]"`
		UserEmail string  `json:"user_email" example:"user@example.com"`
		PromoCode string  `json:"promo_code,omitempty" example:"SUMMER2026"`
		EarlyID   int64   `json:"early_id,omitempty" example:"101"`
		BundleID  int64   `json:"bundle_id,omitempty" example:"202"`
	}

	CreateBookingResponse struct {
		PaymentURL string `json:"payment_url" example:"https://checkout.stripe.com/pay/cs_test_..."`
	}

	OrderByWalletResponse struct {
		BookingID string `json:"booking_id" example:"f9b0b2c0-4e7a-4d1b-9b5e-8c7b4e4f9b12"`
		Status    string `json:"status" example:"CHARGE_STATUS_SUCCESS"`
	}
)

type (
	GetTicketsRequest struct {
		EventID int64 `json:"event_id" example:"1024"`
	}
	GetTicketsResponse struct {
		Tickets []TicketResponse `json:"tickets"`
		Count   int64            `json:"count" example:"1"`
	}
)

type (
	GetEventDiscountRequest struct {
		EventID int64 `json:"event_id" example:"1024"`
	}

	GetEventDiscountResponse struct {
		Early   []EarlyResponse  `json:"early"`
		Bundles []BundleResponse `json:"bundles"`
	}
)

type TicketTransferRequest struct {
	TicketID int64 `json:"ticket_id" example:"5501"`
	ToUserID int64 `json:"to_user_id" example:"123"`
}

type UserTicketsResponse struct {
	Tickets []TicketResponse `json:"tickets"`
}

type TicketResponse struct {
	ID         int64  `json:"id" example:"5501"`
	EventID    int64  `json:"event_id" example:"1024"`
	VenueID    int64  `json:"venue_id" example:"202"`
	SectorName string `json:"sector_name" example:"Platinum A"`
	RowNumber  int64  `json:"row_no" example:"12"`
	SeatNumber int64  `json:"seat_no" example:"42"`
	Price      int64  `json:"price" example:"15000"`
	Status     string `json:"status" example:"created"`
}

type EarlyResponse struct {
	ID         int64     `json:"id" example:"101"`
	EventID    int64     `json:"event_id" example:"1024"`
	Code       string    `json:"code" example:"EARLY2026"`
	Type       string    `json:"type" example:"PERCENT"`
	Value      float64   `json:"value" example:"10"`
	SectorName string    `json:"sector_name" example:"A"`
	ValidUntil time.Time `json:"valid_until" example:"2026-03-01T00:00:00Z"`
}

type BundleResponse struct {
	ID             int64  `json:"id" example:"202"`
	EventID        int64  `json:"event_id" example:"1024"`
	Code           string `json:"code" example:"BUNDLE2+1"`
	SectorName     string `json:"sector_name" example:"B"`
	BundleBuyCount int64  `json:"bundle_buy_count" example:"2"`
	BundleGetCount int64  `json:"bundle_get_count" example:"1"`
}

type RefundBookingRequest struct {
	EventID  int64 `json:"event_id" example:"1024"`
	TicketID int64 `json:"ticket_id" example:"5501"`
}

type WaitlistRequest struct {
	EventID int64  `json:"event_id" example:"1024"`
	Email   string `json:"email" example:"user@example.com"`
}
