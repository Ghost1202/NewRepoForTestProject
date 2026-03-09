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

// AddToWaitlist godoc
// @Summary      Add user to waitlist
// @Description  Adds the current user to the waitlist. User must provide email for notifications.
// @Tags         user
// @Accept       json
// @Success      204 "No Content"
// @Failure      400 {string} string "Bad Request"
// @Failure      401 {string} string "Unauthorized"
// @Router       /user/waitlist [post]
func (h *Handler) AddToWaitlist(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.AddToWaitlist"
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

	var req WaitlistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode body failed")
		log.Error("failed to decode waitlist request body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.Int64("user_id", userID),
		attribute.Int64("event_id", req.EventID),
		attribute.String("user_email", req.Email),
	)

	waitlistDomain := h.reqConv.ToWaitlist(userID, req)
	if err := h.srv.AddToWaitlist(ctx, waitlistDomain); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "service execution failed")
		log.Error("failed to add user to waitlist",
			zap.Error(err),
			zap.Int64("user_id", userID),
			zap.Int64("event_id", req.EventID),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Info("user added to waitlist successfully",
		zap.Int64("user_id", userID),
		zap.Int64("event_id", req.EventID),
	)

	w.WriteHeader(http.StatusNoContent)
}
