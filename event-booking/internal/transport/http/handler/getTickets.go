package handler

import (
	"encoding/json"
	"net/http"

	"github.com/turtlepavlo/event-booking/pkg/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// GetTickets godoc
// @Summary      Get event tickets
// @Description  Get a list of tickets for a specific event from the Slave DB.
// @Tags         tickets
// @Accept       json
// @Produce      json
// @Router       /tickets [get]\n// @Router       /tickets [post]
func (h *Handler) GetTickets(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.GetTickets"
	tracer := otel.Tracer("event-booking/internal/transport/http/handler")

	ctx, span := tracer.Start(r.Context(), op,
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, h.log)
	var request GetTicketsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode body failed")

		log.Error("failed to decode get tickets request body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.Int64("query.event_id", request.EventID))
	tickets, err := h.srv.GetTickets(ctx, request.EventID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "service execution failed")

		log.Error("failed to get tickets",
			zap.Error(err),
			zap.Int64("event_id", request.EventID),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := h.respConv.ToTickets(tickets)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		span.RecordError(err)
		log.Error("failed to encode response", zap.Error(err))
		return
	}

	log.Debug("get tickets request processed", zap.Int64("count", int64(len(tickets))))
}
