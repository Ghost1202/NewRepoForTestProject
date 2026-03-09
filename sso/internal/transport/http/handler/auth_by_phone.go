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

// AuthByPhone godoc
// @Summary      Authorization by phone (get JWT)
// @Description  Validates phone/password and returns JWT access token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      LoginByPhoneReq   true  "Phone credentials"
// @Success      200    {object}  LoginResp
// @Failure      400   "Bad Request"
// @Failure      401   "Unauthorized"
// @Failure      500   "Internal Server Error"
// @Router       /users/auth/phone [post]
func (hand *Handler) AuthByPhone(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.AuthByPhone"
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
		zap.String("route", "/users/auth/phone"),
	)

	var req LoginByPhoneReq
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode body failed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req.Phone = strings.TrimSpace(req.Phone)
	req.Password = strings.TrimSpace(req.Password)
	session, err := hand.reqConverter.AuthByPhone(ctx, hand.authService, req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			span.SetStatus(codes.Error, "invalid credentials")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "service auth by phone failed")
		log.Error("service auth by phone failed", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
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
