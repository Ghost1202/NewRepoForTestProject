package handler

import (
	"encoding/json"
	"net/http"

	"github.com/turtlepavlo/event-service/internal/domain"
	"github.com/turtlepavlo/event-service/internal/lib/telemetry"
	"github.com/turtlepavlo/event-service/internal/transport/http/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// UpdateEvent godoc
// @Summary      Update event details
// @Description  Updates an existing event. Restricted to users with the 'admin' role.
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        request body handler.UpdateEventReq true "Event update data"
// @Success      204 "No Content"
// @Failure      400 {string} string "Bad Request"
// @Failure      401 {string} string "Unauthorized"
// @Security     ApiKeyAuth
// @Router       /events [put]
func (h *Handler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("event-service/internal/transport/http/handler")
	ctx, span := tracer.Start(r.Context(), "Handler.Update")
	defer span.End()

	actor, ok := ctx.Value(middleware.PerformerContextKey{}).(domain.Performer)
	if !ok {
		telemetry.WithTrace(ctx, h.log).Error("performer context missing")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var request UpdateEventReq
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		span.RecordError(err)
		telemetry.WithTrace(ctx, h.log).Warn("invalid update request body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.Int64("event.id", request.EventID))

	event := h.reqConv.ToDomainUpdateEvent(request)

	if err := h.srv.UpdateEvent(ctx, actor, event); err != nil {
		span.SetStatus(codes.Error, "service update failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	telemetry.WithTrace(ctx, h.log).Info("event updated successfully",
		zap.Int64("event_id", request.EventID),
	)

	w.WriteHeader(http.StatusNoContent)
}
