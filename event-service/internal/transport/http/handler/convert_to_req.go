package handler

import (
	"github.com/turtlepavlo/event-service/internal/domain"
)

type RequestConverter struct{}

func NewRequestConverter() *RequestConverter {
	return &RequestConverter{}
}

func (c *RequestConverter) ToDomainCreateEvent(req CreateEventReq) *domain.Event {
	return &domain.Event{
		VenueID:       req.VenueID,
		Name:          req.Name,
		StartDate:     req.DateStart,
		PostDate:      req.PostDate,
		SaleStartDate: req.SaleStartDate,
		MaxPriceCof:   req.MaxPriceCof,
		MinPriceCof:   req.MinPriceCof,
	}
}

func (c *RequestConverter) ToDomainUpdateEvent(req UpdateEventReq) *domain.Event {
	return &domain.Event{
		ID:          req.EventID,
		Name:        req.Name,
		StartDate:   req.DateStart,
		MaxPriceCof: req.MaxPriceCof,
		MinPriceCof: req.MinPriceCof,
	}
}

func (c *RequestConverter) ToDomainTicketSectors(constructor []TicketConstructor) []domain.TicketSector {
	if len(constructor) == 0 {
		return nil
	}
	sectors := make([]domain.TicketSector, 0, len(constructor))
	for _, item := range constructor {
		sectors = append(sectors, domain.TicketSector{
			Name:        item.Name,
			Type:        item.Type,
			Price:       item.Price,
			RowsCount:   item.RowsCount,
			SeatsPerRow: item.SeatsPerRow,
		})
	}
	return sectors
}

func (c *RequestConverter) ToDomainPromos(dtos []Promo, eventID int64) []domain.Promo {
	if len(dtos) == 0 {
		return nil
	}
	result := make([]domain.Promo, len(dtos))
	for i := range dtos {
		result[i] = domain.Promo{
			EventID: eventID,
			Code:    dtos[i].Code,
			Type:    dtos[i].Type,
			Value:   dtos[i].Value,
			Sector:  dtos[i].Sector,
		}
	}

	return result
}

func (c *RequestConverter) ToDomainEarly(dtos []Early, eventID int64) []domain.Early {
	if len(dtos) == 0 {
		return nil
	}
	result := make([]domain.Early, len(dtos))
	for i := range dtos {
		result[i] = domain.Early{
			EventID:    eventID,
			Code:       dtos[i].Code,
			Type:       dtos[i].Type,
			Value:      dtos[i].Value,
			Sector:     dtos[i].Sector,
			ValidUntil: dtos[i].ValidUntil,
		}
	}
	return result
}

func (c *RequestConverter) ToDomainBundles(dtos []Bundle, eventID int64) []domain.Bundle {
	if len(dtos) == 0 {
		return nil
	}
	result := make([]domain.Bundle, len(dtos))
	for i := range dtos {
		result[i] = domain.Bundle{
			EventID:  eventID,
			Code:     dtos[i].Code,
			Sector:   dtos[i].Sector,
			BuyCount: dtos[i].BuyCount,
			GetCount: dtos[i].GetCount,
		}
	}
	return result
}

func (c *RequestConverter) ToDomainFilter(eventID, limit, offset int64) domain.Filter {
	return domain.Filter{
		EventID: eventID,
		Limit:   limit,
		Offset:  offset,
	}
}
