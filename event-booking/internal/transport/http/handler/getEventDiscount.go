package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/turtlepavlo/event-booking/internal/service"
	"github.com/turtlepavlo/event-booking/pkg/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// GetEventDiscount godoc
// @Summary      Get event discount
// @Description  Get early and bundle discounts for a specific event.
// @Tags         discounts
// @Accept       json
// @Produce      json
// @Router       /tickets/discounts [get]\n// @Router       /tickets/discounts [post]
func (h *Handler) GetEventDiscount(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.GetEventDiscount"
	tracer := otel.Tracer("event-booking/internal/transport/http/handler")

	ctx, span := tracer.Start(r.Context(), op,
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, h.log)
	var request GetEventDiscountRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode body failed")

		log.Error("failed to decode get event discount request body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.Int64("query.event_id", request.EventID))
	early, bundles, err := h.srv.GetEventDiscount(ctx, request.EventID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "service execution failed")

		log.Error("failed to get event discounts",
			zap.Error(err),
			zap.Int64("event_id", request.EventID),
		)
		if errors.Is(err, service.ErrInvalidBookingInput) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	response := h.respConv.ToEventDiscount(early, bundles)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		span.RecordError(err)
		log.Error("failed to encode response", zap.Error(err))
		return
	}

	log.Debug("get event discount request processed",
		zap.Int64("early_count", int64(len(early))),
		zap.Int64("bundle_count", int64(len(bundles))),
	)
}
