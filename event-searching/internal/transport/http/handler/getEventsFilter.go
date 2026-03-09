package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/turtlepavlo/event-searching/internal/lib/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// GetEventsFilter godoc
// @Summary      Search events
// @Description  Full-text search for events using ElasticSearch.
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        query         query  string  false  "Search text (name, description)"
// @Param        performer_id  query  int     false  "Filter by performer ID"
// @Param        venue_id      query  int     false  "Filter by venue ID"
// @Param        date_from     query  string  false  "Start date (RFC3339)" Format(date-time)
// @Param        date_to       query  string  false  "End date (RFC3339)" Format(date-time)
// @Param        limit         query  int     false  "Max records to return"
// @Param        offset        query  int     false  "Number of records to skip"
// @Success      200           {object}  GetEventsFilterResponse "Search results"
// @Failure      400           {string}  string  "Bad Request"
// @Failure      500           {string}  string  "Internal Server Error"
// @Router       /events/search [get]
func (h *Handler) GetEventsFilter(write http.ResponseWriter, req *http.Request) {
	const op = "Handler.GetEventsFilter"
	var tracer = otel.Tracer("http/handler")

	ctx, span := tracer.Start(req.Context(), op,
		trace.WithAttributes(
			attribute.String("http.method", req.Method),
			attribute.String("http.url", req.URL.String()),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, h.log)

	var request GetEventsFilterRequest
	var decoder = schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	if err := decoder.Decode(&request, req.URL.Query()); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode params failed")

		log.Error("failed to decode search params", zap.Error(err))
		write.WriteHeader(http.StatusBadRequest)
		return
	}

	attrs := make([]attribute.KeyValue, 0, 7)
	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}

	domainFilter := h.reqConv.ToDomainFilter(request)
	events, err := h.srv.SearchEventsFilter(ctx, domainFilter)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "search service execution failed")

		log.Error("failed to search events", zap.Error(err))
		write.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := h.respConv.ToEventsFilterResponse(events, int64(len(events)))
	write.Header().Set("Content-Type", "application/json")
	write.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(write).Encode(response); err != nil {
		span.RecordError(err)
		log.Error("failed to encode search response", zap.Error(err))
		return
	}

	log.Debug("search request processed", zap.Int("count", len(events)))
}
