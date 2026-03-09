package handler

import "github.com/turtlepavlo/event-searching/internal/domain"

type RequestConverter struct{}

func NewRequestConverter() *RequestConverter { return &RequestConverter{} }

func (c *RequestConverter) ToDomainFilter(req GetEventsFilterRequest) domain.SearchFilter {
	return domain.SearchFilter{
		Query:       req.Query,
		PerformerID: req.PerformerID,
		VenueID:     req.VenueID,
		DateFrom:    req.DateFrom,
		DateTo:      req.DateTo,
		Limit:       req.Limit,
		Offset:      req.Offset,
	}
}

func (c *RequestConverter) ToDomainUpsertComment(req SetCommentRequest, userID int64) domain.UpsertComment {
	return domain.UpsertComment{
		EventID: req.EventID,
		UserID:  userID,
		Nick:    req.Nick,
		Text:    req.Text,
		Rating:  req.Rating,
	}
}

func (c *RequestConverter) ToDomainListComments(req GetCommentsRequest) domain.ListComments {
	return domain.ListComments{
		EventID: req.EventID,
		Limit:   req.Limit,
		Offset:  req.Offset,
	}
}

func (c *RequestConverter) ToDomainEventID(req GetRatingRequest) int64 {
	return req.EventID
}
