package handler

import "github.com/turtlepavlo/event-searching/internal/domain"

type ResponseConverter struct{}

func NewResponseConverter() *ResponseConverter { return &ResponseConverter{} }

func (c *ResponseConverter) ToEventResponse(e domain.Event) EventResponse {
	return EventResponse{
		ID:            e.ID,
		PerformerID:   e.PerformerID,
		VenueID:       e.VenueID,
		Name:          e.Name,
		StartDate:     e.StartDate,
		SoldOut:       e.SoldOut,
		SaleStartDate: e.SaleStartDate,
		InfoHeader:    e.InfoHeader,
		InfoBody:      e.InfoBody,
		Popularity:    e.Popularity,
	}
}

func (c *ResponseConverter) ToEventsResponse(events []domain.Event, count int64) GetEventsResponse {
	res := make([]EventResponse, len(events))
	for i := range events {
		res[i] = c.ToEventResponse(events[i])
	}
	return GetEventsResponse{
		Events: res,
		Count:  count,
	}
}

func (c *ResponseConverter) ToEventsFilterResponse(events []domain.Event, count int64) GetEventsFilterResponse {
	res := make([]EventResponse, len(events))
	for i := range events {
		res[i] = c.ToEventResponse(events[i])
	}
	return GetEventsFilterResponse{
		Events: res,
		Count:  count,
	}
}

func (c *ResponseConverter) ToComment(item domain.Comment) Comment {
	return Comment{
		EventID: item.EventID,
		UserID:  item.UserID,
		Nick:    item.Nick,
		Text:    item.Text,
		Rating:  item.Rating,
	}
}

func (c *ResponseConverter) ToCommentsResponse(req GetCommentsRequest, items []domain.Comment) GetCommentsResponse {
	out := make([]Comment, len(items))
	for i := range items {
		out[i] = c.ToComment(items[i])
	}

	return GetCommentsResponse{
		EventID: req.EventID,
		Limit:   req.Limit,
		Offset:  req.Offset,
		Items:   out,
	}
}

func (c *ResponseConverter) ToRatingResponse(r domain.Rating) GetRatingResponse {
	return GetRatingResponse{
		EventID: r.EventID,
		Avg:     r.Avg,
		Count:   r.Count,
	}
}
