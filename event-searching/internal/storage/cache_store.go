package storage

import (
	"context"
	"encoding/json"

	"github.com/olivere/elastic/v7"
	"github.com/turtlepavlo/event-searching/internal/lib/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type FavouriteStore struct {
	elastic *elastic.Client
	log     *zap.Logger
}

func NewFavouriteStore(el *elastic.Client, log *zap.Logger) *FavouriteStore {
	return &FavouriteStore{
		elastic: el,
		log:     log,
	}
}

func (repo *FavouriteStore) Search(ctx context.Context, filter SearchFilterDTO) ([]EventModel, error) {
	const op = "repository.Query.Search"
	tracer := otel.Tracer("repository")
	ctx, span := tracer.Start(ctx, op)
	defer span.End()

	logger := telemetry.WithTrace(ctx, repo.log)

	query := elastic.NewBoolQuery()

	if filter.Query != "" {
		query.Must(elastic.NewMultiMatchQuery(filter.Query, "name", "info_header", "info_body").
			Fuzziness("AUTO").
			Type("best_fields"))
	}

	if filter.PerformerID != 0 {
		query.Filter(elastic.NewTermQuery("performer_id", filter.PerformerID))
	}
	if filter.VenueID != 0 {
		query.Filter(elastic.NewTermQuery("venue_id", filter.VenueID))
	}

	if !filter.DateFrom.IsZero() || !filter.DateTo.IsZero() {
		rangeQ := elastic.NewRangeQuery("date_start")
		if !filter.DateFrom.IsZero() {
			rangeQ.Gte(filter.DateFrom)
		}
		if !filter.DateTo.IsZero() {
			rangeQ.Lte(filter.DateTo)
		}
		query.Filter(rangeQ)
	}

	searchResult, err := repo.elastic.Search().
		Index("event_service.public.events").
		Query(query).
		From(int(filter.Offset)).
		Size(int(filter.Limit)).
		Do(ctx)

	if err != nil {
		if elastic.IsNotFound(err) {
			return []EventModel{}, nil
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "elastic search failed")
		logger.Error("elastic search request failed", zap.Error(err))
		return nil, err
	}

	events := make([]EventModel, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		var doc EventModel
		if err := json.Unmarshal(hit.Source, &doc); err != nil {
			logger.Error("failed to unmarshal elastic doc", zap.String("id", hit.Id))
			continue
		}
		events = append(events, doc)
	}

	return events, nil
}
