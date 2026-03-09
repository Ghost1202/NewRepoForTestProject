package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/turtlepavlo/event-booking/internal/service"
	"github.com/turtlepavlo/event-booking/internal/transport/http/middleware"
	"github.com/turtlepavlo/event-booking/pkg/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// RefundBooking godoc
// @Summary      Refund a booked ticket
// @Description  Starts refund flow and returns refund URL.
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Router 		 /api/v1/bookings/refund [post]
// @Security     BearerAuth
func (h *Handler) RefundBooking(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.RefundBooking"
	tracer := otel.Tracer("/http/handler")

	ctx, span := tracer.Start(r.Context(), op,
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, h.log)

	var req RefundBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode body failed")
		log.Error("failed to decode request body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, ok := ctx.Value(middleware.UserContextKey{}).(int64)
	if !ok {
		span.SetStatus(codes.Error, "user_id missing in context")
		log.Error("unauthorized: user_id missing in context")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	refund := h.reqConv.ToRefund(userID, req)
	span.SetAttributes(
		attribute.Int64("user_id", refund.UserID),
		attribute.Int64("event_id", refund.EventID),
		attribute.Int64("ticket_id", refund.TicketID),
	)

	if err := h.srv.RefundTicket(ctx, refund); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "refund failed")
		log.Error("refund failed",
			zap.Error(err),
			zap.Int64("user_id", refund.UserID),
			zap.Int64("event_id", refund.EventID),
			zap.Int64("ticket_id", refund.TicketID),
		)

		switch {
		case errors.Is(err, service.ErrTicketNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, service.ErrTicketNotOwned):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, service.ErrBookingNotPaid):
			w.WriteHeader(http.StatusPreconditionFailed)
		case errors.Is(err, service.ErrRefundAlreadyProcessed):
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
