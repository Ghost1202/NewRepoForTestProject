package service

import (
	"github.com/turtlepavlo/event-booking/internal/domain"
	"github.com/turtlepavlo/event-booking/internal/storage"
)

type FromStorageConvert struct{}

func NewFromStorageConvert() *FromStorageConvert { return &FromStorageConvert{} }

func (c *FromStorageConvert) Tickets(models []storage.TicketModel) []domain.Ticket {
	tickets := make([]domain.Ticket, len(models))
	for i := range models {
		tickets[i] = c.Ticket(&models[i])
	}
	return tickets
}

func (c *FromStorageConvert) Ticket(dto *storage.TicketModel) domain.Ticket {
	return domain.Ticket{
		ID:         dto.ID,
		EventID:    dto.EventID,
		VenueID:    dto.VenueID,
		SectorName: dto.SectorName,
		RowNumber:  dto.RowNumber,
		SeatNumber: dto.SeatNumber,
		Price:      dto.Price,
		Status:     dto.Status,
		UserID:     dto.UserID,
		CreatedAt:  dto.CreatedAt,
	}
}

func (c *FromStorageConvert) Promo(promo *storage.Promo) *domain.PromoCode {
	if promo == nil {
		return nil
	}
	return &domain.PromoCode{
		ID:      promo.ID,
		EventID: promo.EventID,
		Code:    promo.Code,
		Type:    promo.Type,
		Value:   promo.Value,
		Sector:  promo.SectorName,
	}
}

func (c *FromStorageConvert) Earlies(models []storage.Early) []domain.Early {
	out := make([]domain.Early, len(models))
	for i := range models {
		out[i] = c.Early(&models[i])
	}
	return out
}

func (c *FromStorageConvert) Early(model *storage.Early) domain.Early {
	return domain.Early{
		ID:         model.ID,
		EventID:    model.EventID,
		Code:       model.Code,
		Type:       model.Type,
		Value:      model.Value,
		SectorName: model.SectorName,
		ValidUntil: model.ValidUntil,
	}
}

func (c *FromStorageConvert) Bundles(models []storage.Bundle) []domain.Bundle {
	out := make([]domain.Bundle, len(models))
	for i := range models {
		out[i] = c.Bundle(&models[i])
	}
	return out
}

func (c *FromStorageConvert) Bundle(model *storage.Bundle) domain.Bundle {
	return domain.Bundle{
		ID:             model.ID,
		EventID:        model.EventID,
		Code:           model.Code,
		SectorName:     model.SectorName,
		BundleBuyCount: model.BundleBuyCount,
		BundleGetCount: model.BundleGetCount,
	}
}
