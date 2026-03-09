package service

import (
	"github.com/turtlepavlo/event-service/internal/domain"
	"github.com/turtlepavlo/event-service/internal/storage"
)

type converte struct{}

func (c converte) toRepoEventModel(event *domain.Event) storage.Events {
	return storage.Events{
		ID:            event.ID,
		PerformerID:   event.PerformerID,
		VenueID:       event.VenueID,
		Name:          event.Name,
		DateStart:     event.StartDate,
		SoldOut:       event.SoldOut,
		PostDate:      event.PostDate,
		SaleStartDate: event.SaleStartDate,
		MaxPriceCof:   event.MaxPriceCof,
		MinPriceCof:   event.MinPriceCof,
	}
}

func (c converte) toRepoTickets(venueID int64, sectors []domain.TicketSector) []storage.Tickets {
	var count int
	for _, s := range sectors {
		count += int(s.RowsCount * s.SeatsPerRow)
	}
	tickets := make([]storage.Tickets, count)
	var i int
	for _, s := range sectors {
		base := storage.Tickets{
			VenueID:    venueID,
			SectorName: s.Name,
			Price:      s.Price,
			Status:     StatusCreated,
			UserID:     0,
		}
		rows := int(s.RowsCount)
		seats := int(s.SeatsPerRow)
		for r := 0; r < rows; r++ {
			rowNum := int64(r + 1)
			for st := 0; st < seats; st++ {
				tickets[i] = base
				tickets[i].RowNumber = rowNum
				tickets[i].SeatNumber = int64(st + 1)
				i++
			}
		}
	}
	return tickets
}

func (c converte) ToStorageEvent(d *domain.Event) storage.Events {
	return storage.Events{
		ID:            d.ID,
		PerformerID:   d.PerformerID,
		VenueID:       d.VenueID,
		Name:          d.Name,
		DateStart:     d.StartDate,
		PostDate:      d.PostDate,
		SaleStartDate: d.SaleStartDate,
		MaxPriceCof:   d.MaxPriceCof,
		MinPriceCof:   d.MinPriceCof,
	}
}

func (c converte) ToStoragePromos(domains []domain.Promo) []storage.Promo {
	res := make([]storage.Promo, len(domains))
	for i := range domains {
		res[i] = storage.Promo{
			ID:         domains[i].ID,
			EventID:    domains[i].EventID,
			Code:       domains[i].Code,
			Type:       domains[i].Type,
			Value:      domains[i].Value,
			SectorName: domains[i].Sector,
		}
	}
	return res
}

func (c converte) ToStoragePromosWithEvent(domains []domain.Promo, eventID int64) []storage.Promo {
	res := make([]storage.Promo, len(domains))
	for i := range domains {
		id := domains[i].EventID
		if eventID != 0 {
			id = eventID
		}
		res[i] = storage.Promo{
			ID:         domains[i].ID,
			EventID:    id,
			Code:       domains[i].Code,
			Type:       domains[i].Type,
			Value:      domains[i].Value,
			SectorName: domains[i].Sector,
		}
	}
	return res
}

func (c converte) ToStorageEarly(domains []domain.Early) []storage.Early {
	res := make([]storage.Early, len(domains))
	for i := range domains {
		res[i] = storage.Early{
			ID:         domains[i].ID,
			EventID:    domains[i].EventID,
			Code:       domains[i].Code,
			Type:       domains[i].Type,
			Value:      domains[i].Value,
			SectorName: domains[i].Sector,
			ValidUntil: domains[i].ValidUntil,
		}
	}
	return res
}

func (c converte) ToStorageEarlyWithEvent(domains []domain.Early, eventID int64) []storage.Early {
	res := make([]storage.Early, len(domains))
	for i := range domains {
		id := domains[i].EventID
		if eventID != 0 {
			id = eventID
		}
		res[i] = storage.Early{
			ID:         domains[i].ID,
			EventID:    id,
			Code:       domains[i].Code,
			Type:       domains[i].Type,
			Value:      domains[i].Value,
			SectorName: domains[i].Sector,
			ValidUntil: domains[i].ValidUntil,
		}
	}
	return res
}

func (c converte) ToStorageBundles(domains []domain.Bundle) []storage.Bundle {
	res := make([]storage.Bundle, len(domains))
	for i := range domains {
		res[i] = storage.Bundle{
			ID:         domains[i].ID,
			EventID:    domains[i].EventID,
			Code:       domains[i].Code,
			SectorName: domains[i].Sector,
			BuyCount:   domains[i].BuyCount,
			GetCount:   domains[i].GetCount,
		}
	}
	return res
}

func (c converte) ToStorageBundlesWithEvent(domains []domain.Bundle, eventID int64) []storage.Bundle {
	res := make([]storage.Bundle, len(domains))
	for i := range domains {
		id := domains[i].EventID
		if eventID != 0 {
			id = eventID
		}
		res[i] = storage.Bundle{
			ID:         domains[i].ID,
			EventID:    id,
			Code:       domains[i].Code,
			SectorName: domains[i].Sector,
			BuyCount:   domains[i].BuyCount,
			GetCount:   domains[i].GetCount,
		}
	}
	return res
}

func (c converte) ToStorageTicketTransfer(d domain.TicketTransfer) storage.TicketTransfer {
	return storage.TicketTransfer{
		TicketID:   d.TicketID,
		FromUserID: d.FromUserID,
		ToUserID:   d.ToUserID,
	}
}

func (c converte) ToDomainPromos(st []storage.Promo) []domain.Promo {
	if len(st) == 0 {
		return []domain.Promo{}
	}
	res := make([]domain.Promo, len(st))
	for i := range st {
		res[i] = domain.Promo{
			ID:      st[i].ID,
			EventID: st[i].EventID,
			Code:    st[i].Code,
			Type:    st[i].Type,
			Value:   st[i].Value,
			Sector:  st[i].SectorName,
		}
	}
	return res
}

func (c converte) ToDomainEarly(st []storage.Early) []domain.Early {
	if len(st) == 0 {
		return []domain.Early{}
	}
	res := make([]domain.Early, len(st))
	for i := range st {
		res[i] = domain.Early{
			ID:         st[i].ID,
			EventID:    st[i].EventID,
			Code:       st[i].Code,
			Type:       st[i].Type,
			Value:      st[i].Value,
			Sector:     st[i].SectorName,
			ValidUntil: st[i].ValidUntil,
		}
	}
	return res
}

func (c converte) ToDomainBundles(st []storage.Bundle) []domain.Bundle {
	if len(st) == 0 {
		return []domain.Bundle{}
	}
	res := make([]domain.Bundle, len(st))
	for i := range st {
		res[i] = domain.Bundle{
			ID:       st[i].ID,
			EventID:  st[i].EventID,
			Code:     st[i].Code,
			Sector:   st[i].SectorName,
			BuyCount: st[i].BuyCount,
			GetCount: st[i].GetCount,
		}
	}
	return res
}
