package handler

import "github.com/turtlepavlo/event-booking/internal/domain"

type ResponseConverter struct{}

func NewResponseConverter() *ResponseConverter { return &ResponseConverter{} }

func (c *ResponseConverter) ToTicket(t domain.Ticket) TicketResponse {
	return TicketResponse{
		ID:         t.ID,
		EventID:    t.EventID,
		VenueID:    t.VenueID,
		SectorName: t.SectorName,
		RowNumber:  t.RowNumber,
		SeatNumber: t.SeatNumber,
		Price:      t.Price,
		Status:     t.Status,
	}
}

func (c *ResponseConverter) ToUserTickets(tickets []domain.Ticket) UserTicketsResponse {
	res := make([]TicketResponse, len(tickets))
	for i := range tickets {
		res[i] = c.ToTicket(tickets[i])
	}
	return UserTicketsResponse{
		Tickets: res,
	}
}

func (c *ResponseConverter) ToTickets(tickets []domain.Ticket) GetTicketsResponse {
	res := make([]TicketResponse, len(tickets))
	for i := range tickets {
		res[i] = c.ToTicket(tickets[i])
	}
	return GetTicketsResponse{
		Tickets: res,
		Count:   int64(len(res)),
	}
}

func (c *ResponseConverter) ToCreateBooking(paymentURL string) CreateBookingResponse {
	return CreateBookingResponse{
		PaymentURL: paymentURL,
	}
}

func (c *ResponseConverter) ToOrderByWallet(result domain.WalletChargeResult) OrderByWalletResponse {
	return OrderByWalletResponse{
		BookingID: result.BookingID,
		Status:    string(result.Status),
	}
}

func (c *ResponseConverter) ToEventDiscount(early []domain.Early, bundles []domain.Bundle) GetEventDiscountResponse {
	earlyRes := make([]EarlyResponse, len(early))
	for i := range early {
		earlyRes[i] = c.ToEarly(early[i])
	}

	bundleRes := make([]BundleResponse, len(bundles))
	for i := range bundles {
		bundleRes[i] = c.ToBundle(bundles[i])
	}

	return GetEventDiscountResponse{
		Early:   earlyRes,
		Bundles: bundleRes,
	}
}

func (c *ResponseConverter) ToEarly(early domain.Early) EarlyResponse {
	return EarlyResponse{
		ID:         early.ID,
		EventID:    early.EventID,
		Code:       early.Code,
		Type:       early.Type,
		Value:      early.Value,
		SectorName: early.SectorName,
		ValidUntil: early.ValidUntil,
	}
}

func (c *ResponseConverter) ToBundle(bundle domain.Bundle) BundleResponse {
	return BundleResponse{
		ID:             bundle.ID,
		EventID:        bundle.EventID,
		Code:           bundle.Code,
		SectorName:     bundle.SectorName,
		BundleBuyCount: bundle.BundleBuyCount,
		BundleGetCount: bundle.BundleGetCount,
	}
}
