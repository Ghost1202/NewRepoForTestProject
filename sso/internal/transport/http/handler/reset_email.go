package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	telemetry "github.com/turtlepavlo/sso/pkg/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// ResetEmail godoc
// @Summary      Reset password by email (send OTP)
// @Description  Sends OTP to email for password reset.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      ResetEmailReq  true  "Email"
// @Success      204
// @Failure      400
// @Failure      500
// @Router       /users/reset/email [post]
func (hand *Handler) ResetEmail(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.ResetEmail"
	tracer := otel.Tracer("/http/handler")

	ctx, span := tracer.Start(r.Context(), op,
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, hand.log).With(
		zap.String("op", op),
		zap.String("route", "/users/reset/email"),
	)

	var req ResetEmailReq
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode body failed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		span.SetStatus(codes.Error, "missing email")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("auth.destination_type", "email"),
		attribute.String("user.email", req.Email),
	)

	if err := hand.reqConverter.ResetEmail(ctx, hand.authService, req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "send otp failed")
		log.Error("send reset otp failed", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
