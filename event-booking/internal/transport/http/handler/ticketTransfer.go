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

// TicketTransfer godoc
// @Summary      Ticket transfer
// @Description  Publishes ticket transfer request.
// @Tags         tickets
// @Accept       json
// @Success      200
// @Router       /user/tickets/transfer [post]
func (h *Handler) TicketTransfer(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.TicketTransfer"
	tracer := otel.Tracer("/http/handler")

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

	var req TicketTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode body failed")
		log.Error("failed to decode request body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	transfer := h.reqConv.ToTicketTransfer(userID, req)
	span.SetAttributes(
		attribute.Int64("user_id", transfer.FromUserID),
		attribute.Int64("ticket_id", transfer.TicketID),
		attribute.Int64("to_user_id", transfer.ToUserID),
	)

	if err := h.srv.TransferTicket(ctx, transfer); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "transfer publish failed")
		log.Error("transfer ticket publish failed",
			zap.Error(err),
			zap.Int64("ticket_id", transfer.TicketID),
			zap.Int64("from_user_id", transfer.FromUserID),
			zap.Int64("to_user_id", transfer.ToUserID),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
