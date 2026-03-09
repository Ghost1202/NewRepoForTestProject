package service

import "github.com/turtlepavlo/event-booking/internal/storage"

func (srv *BookingService) isValidPromo(promo storage.Promo, ticket storage.TicketModel) bool {
	if promo.SectorName != "" && promo.SectorName != ticket.SectorName {
		return false
	}
	return true
}
