package service

import (
	"time"

	"github.com/turtlepavlo/event-searching/internal/domain"
	"github.com/turtlepavlo/event-searching/internal/storage"
	sqlc "github.com/turtlepavlo/event-searching/internal/storage/postgres/sqlc"
)

type converteToDomain struct{}

func (c converteToDomain) toDomainEvent(doc storage.EventModel) domain.Event {
	return domain.Event{
		ID:            doc.ID,
		PerformerID:   doc.PerformerID,
		VenueID:       doc.VenueID,
		Name:          doc.Name,
		StartDate:     doc.StartDate,
		SoldOut:       doc.SoldOut,
		SaleStartDate: doc.SaleStartDate,
		InfoHeader:    doc.InfoHeader,
		InfoBody:      doc.InfoBody,
		Popularity:    doc.Popularity,
	}
}

func (c converteToDomain) toDomainEvents(docs []storage.EventModel) []domain.Event {
	events := make([]domain.Event, len(docs))
	for i := range docs {
		events[i] = c.toDomainEvent(docs[i])
	}
	return events
}

func (c converteToDomain) toDomainComment(m sqlc.EventComment) domain.Comment {
	var createdAt time.Time
	if m.CreatedAt.Valid {
		createdAt = m.CreatedAt.Time
	}

	return domain.Comment{
		ID:        m.ID,
		EventID:   m.EventID,
		UserID:    m.UserID,
		Nick:      m.Nick,
		Text:      m.Text,
		Rating:    m.RatingTenths,
		CreatedAt: createdAt,
	}
}

func (c converteToDomain) toDomainComments(items []sqlc.EventComment) []domain.Comment {
	out := make([]domain.Comment, len(items))
	for i := range items {
		out[i] = c.toDomainComment(items[i])
	}
	return out
}

func (c converteToDomain) toDomainRating(r sqlc.GetRatingByEventRow) domain.Rating {
	return domain.Rating{
		EventID: r.EventID,
		Avg:     r.AvgTenths,
		Count:   r.Count,
	}
}
