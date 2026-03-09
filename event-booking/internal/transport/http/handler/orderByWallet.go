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

// OrderByWallet godoc
// @Summary      Book tickets using Wallet
// @Description  Reserves tickets in Redis, applies one discount type, charges Wallet via gRPC, and returns charge status.
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Router       /bookings/wallet [post]
// @Security     BearerAuth
func (h *Handler) OrderByWallet(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.OrderByWallet"
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

	var request CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode body failed")
		log.Error("failed to decode request body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, ok := ctx.Value(middleware.UserContextKey{}).(int64)
	if !ok {
		span.SetStatus(codes.Error, "user_id missing in context")
		log.Error("unauthorized access attempt: user_id missing")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	booking := h.reqConv.ToBooking(userID, request)
	span.SetAttributes(
		attribute.Int64("user_id", booking.UserID),
		attribute.Int64("ticket_ids.count", int64(len(booking.TicketIDs))),
		attribute.String("promo_code", booking.PromoCode),
		attribute.Int64("early_id", booking.EarlyID),
		attribute.Int64("bundle_id", booking.BundleID),
	)

	result, err := h.srv.OrderByWallet(ctx, booking)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "wallet booking failed")

		log.Error("wallet booking failed",
			zap.Error(err),
			zap.Int64("user_id", booking.UserID),
			zap.Int64("ticket_ids.count", int64(len(booking.TicketIDs))),
			zap.String("promo_code", booking.PromoCode),
			zap.Int64("early_id", booking.EarlyID),
			zap.Int64("bundle_id", booking.BundleID),
		)

		switch {
		case errors.Is(err, service.ErrTicketAlreadyReserved):
			w.WriteHeader(http.StatusConflict)
		case errors.Is(err, service.ErrTicketNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, service.ErrInvalidBookingInput):
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	statusCode := walletEnum(result.Status)
	if statusCode == http.StatusInternalServerError {
		span.SetStatus(codes.Error, "wallet charge status unspecified")
		log.Error("wallet charge returned unexpected status",
			zap.String("status", string(result.Status)),
			zap.String("booking_id", result.BookingID),
		)
	}

	response := h.respConv.ToOrderByWallet(result)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error("encode response failed", zap.Error(err))
	}
}
