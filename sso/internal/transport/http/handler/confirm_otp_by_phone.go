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

// ConfirmOTPByPhone godoc
// @Summary      Confirm OTP by phone (get JWT)
// @Description  Confirms OTP and returns JWT access token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      ConfirmOTPByPhoneReq  true  "OTP confirm"
// @Success      200    {object}  LoginResp
// @Failure      400   "Bad Request"
// @Failure      401   "Unauthorized"
// @Failure      404   "Not Found"
// @Failure      500   "Internal Server Error"
// @Router       /users/otp/phone/confirm [post]
func (hand *Handler) ConfirmOTPByPhone(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.ConfirmOTPByPhone"
	tr := otel.Tracer("http/handler")

	ctx, span := tr.Start(r.Context(), op,
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, hand.log).With(
		zap.String("op", op),
		zap.String("route", "/users/otp/phone/confirm"),
	)

	var req ConfirmOTPByPhoneReq
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode body failed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req.Phone = strings.TrimSpace(req.Phone)
	if req.Phone == "" || req.Code <= 0 {
		span.SetStatus(codes.Error, "missing required fields")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	session, err := hand.reqConverter.ConfirmOTPByPhone(ctx, hand.authService, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCodeNotFound):
			span.SetStatus(codes.Error, "otp code not found")
			w.WriteHeader(http.StatusNotFound)
			return
		case errors.Is(err, service.ErrInvalidCredentials):
			span.SetStatus(codes.Error, "invalid otp")
			w.WriteHeader(http.StatusUnauthorized)
			return
		default:
			span.RecordError(err)
			span.SetStatus(codes.Error, "service confirm otp failed")
			log.Error("service confirm otp failed", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
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
