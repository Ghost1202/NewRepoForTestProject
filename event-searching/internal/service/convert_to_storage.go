package service

import (
	"github.com/turtlepavlo/event-searching/internal/domain"
	"github.com/turtlepavlo/event-searching/internal/storage"
	sqlc "github.com/turtlepavlo/event-searching/internal/storage/postgres/sqlc"
)

type converterToStorage struct{}

func (converterToStorage) toStorageFilter(filter domain.SearchFilter) storage.SearchFilterDTO {
	return storage.SearchFilterDTO{
		Query:       filter.Query,
		PerformerID: filter.PerformerID,
		VenueID:     filter.VenueID,
		DateFrom:    filter.DateFrom,
		DateTo:      filter.DateTo,
		Limit:       filter.Limit,
		Offset:      filter.Offset,
	}
}

func (converterToStorage) toPopularFilter(limit, offset int64) storage.SearchFilterDTO {
	return storage.SearchFilterDTO{
		Limit:  limit,
		Offset: offset,
	}
}

func (c converterToStorage) toCreateCommentParams(comment domain.UpsertComment) sqlc.CreateCommentParams {
	return sqlc.CreateCommentParams{
		EventID:      comment.EventID,
		UserID:       comment.UserID,
		Nick:         comment.Nick,
		Text:         comment.Text,
		RatingTenths: comment.Rating,
	}
}

func (c converterToStorage) toListCommentsParamsFromParts(eventID, limit, offset int64) sqlc.ListCommentsByEventParams {
	return sqlc.ListCommentsByEventParams{
		EventID:    eventID,
		PageOffset: offset,
		PageLimit:  limit,
	}
}
