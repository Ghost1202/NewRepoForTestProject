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

// GetEvents godoc
// @Summary      Get popular events
// @Description  Get a list of popular events from Redis cache (Hot Data). Ideal for feeds.
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        limit   query  int  false  "Max records to return"
// @Param        offset  query  int  false  "Number of records to skip"
// @Success      200     {object}  GetEventsResponse "List of popular events"
// @Failure      400     {string}  string  "Bad Request"
// @Failure      500     {string}  string  "Internal Server Error"
// @Router       /events [get]
func (h *Handler) GetEvents(write http.ResponseWriter, req *http.Request) {
	const op = "Handler.GetEvents"
	var tracer = otel.Tracer("http/handler")

	ctx, span := tracer.Start(req.Context(), op,
		trace.WithAttributes(
			attribute.String("http.method", req.Method),
			attribute.String("http.url", req.URL.String()),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, h.log)

	var request GetEventsRequest
	var decoder = schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	if err := decoder.Decode(&request, req.URL.Query()); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode params failed")

		log.Error("failed to decode query params", zap.Error(err))
		write.WriteHeader(http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.Int64("query.limit", request.Limit),
		attribute.Int64("query.offset", request.Offset),
	)

	events, err := h.srv.GetPopularEvents(ctx, request.Limit, request.Offset)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "service execution failed")

		log.Error("failed to get popular events", zap.Error(err))
		write.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := h.respConv.ToEventsResponse(events, int64(len(events)))
	write.Header().Set("Content-Type", "application/json")
	write.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(write).Encode(response); err != nil {
		span.RecordError(err)
		log.Error("failed to encode response", zap.Error(err))
		return
	}

	log.Debug("popular events request processed", zap.Int("count", len(events)))
}
