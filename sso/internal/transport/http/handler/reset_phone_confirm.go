package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/turtlepavlo/sso/internal/service"
	telemetry "github.com/turtlepavlo/sso/pkg/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// ResetPhoneConfirm godoc
// @Summary      Confirm phone reset (OTP + new password)
// @Description  Confirms OTP and sets new password, returns JWT.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      ResetPhoneConfirmReq  true  "Confirm reset"
// @Success      200    {object}  LoginResp
// @Failure      400
// @Failure      401
// @Failure      404
// @Failure      500
// @Router       /users/reset/phone/confirm [post]
func (hand *Handler) ResetPhoneConfirm(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.ResetPhoneConfirm"
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
		zap.String("route", "/users/reset/phone/confirm"),
	)

	var req ResetPhoneConfirmReq
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode body failed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req.Phone = strings.TrimSpace(req.Phone)
	req.NewPassword = strings.TrimSpace(req.NewPassword)

	if req.Phone == "" || req.NewPassword == "" || req.Code == 0 {
		span.SetStatus(codes.Error, "missing required fields")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("auth.destination_type", "phone"),
		attribute.String("user.phone", req.Phone),
	)

	session, err := hand.reqConverter.ResetPhoneConfirm(ctx, hand.authService, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "confirm reset failed")

		switch {
		case errors.Is(err, service.ErrCodeNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, service.ErrInvalidCredentials):
			w.WriteHeader(http.StatusUnauthorized)
		default:
			log.Error("confirm phone reset failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	out := hand.respConverter.ToLoginResponse(session)

	span.SetAttributes(
		attribute.Int64("user_id", out.UserID),
		attribute.String("user.role", out.Role),
	)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(out); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "encode response failed")
		log.Error("encode response failed", zap.Error(err))
	}
}
