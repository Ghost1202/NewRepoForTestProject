package handler

import (
	"encoding/json"
	"net/http"

	"github.com/turtlepavlo/event-searching/internal/lib/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// GetRating godoc
// @Summary      Get rating
// @Description  Get aggregated rating for an event. Avg is integer in tenths (0..50); client divides by 10.
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        request  body      GetRatingRequest    true  "rating params"
// @Success      200      {object}  GetRatingResponse   "rating aggregate"
// @Failure      400      {string}  string             "Bad Request"
// @Failure      500      {string}  string             "Internal Server Error"
// @Router       /comments/rating [post]
func (h *Handler) GetRating(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.GetRating"
	var tracer = otel.Tracer("http/handler")

	ctx, span := tracer.Start(r.Context(), op,
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, h.log)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var req GetRatingRequest
	if err := dec.Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode body failed")
		log.Error("failed to decode request body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	span.SetAttributes(attribute.Int64("comment.event_id", req.EventID))

	eventID := h.reqConv.ToDomainEventID(req)
	rating, err := h.srv.GetRating(ctx, eventID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "service execution failed")
		log.Error("failed to get rating", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := h.respConv.ToRatingResponse(rating)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "encode response failed")
		log.Error("failed to encode response", zap.Error(err))
		return
	}
}
