package handler

import (
	"encoding/json"
	"net/http"

	"github.com/turtlepavlo/event-booking/internal/transport/http/middleware"
	"github.com/turtlepavlo/event-booking/pkg/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// GetUserTickets godoc
// @Summary      Get user tickets
// @Description  Returns all tickets owned by current user.
// @Tags         tickets
// @Produce      json
// @Success      200 {object} UserTicketsResponse
// @Router       /user/tickets [get]\n// @Router       /user/tickets [post]
func (h *Handler) GetUserTickets(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.GetUserTickets"
	tracer := otel.Tracer("http/handler")

	ctx, span := tracer.Start(r.Context(), op,
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
		),
	)
	defer span.End()
	defer r.Body.Close()

	log := telemetry.WithTrace(ctx, h.log)

	userID, ok := ctx.Value(middleware.UserContextKey{}).(int64)
	if !ok {
		span.SetStatus(codes.Error, "user_id missing in context")
		log.Error("unauthorized access attempt: user_id missing")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	span.SetAttributes(attribute.Int64("user_id", userID))

	tickets, err := h.srv.GetUserTickets(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "get user tickets failed")
		log.Error("get user tickets failed",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := h.respConv.ToUserTickets(tickets)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if encErr := json.NewEncoder(w).Encode(resp); encErr != nil {
		log.Error("encode response failed", zap.Error(encErr))
	}
}
