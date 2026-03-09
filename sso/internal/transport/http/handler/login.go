package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/turtlepavlo/sso/internal/domain"
	telemetry "github.com/turtlepavlo/sso/pkg/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Login godoc
// @Summary      JWT verification / user info
// @Description  Protected endpoint. Returns user data extracted from the Bearer token.
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  AuthResp
// @Failure      400   "Bad Request"
// @Failure      401   "Unauthorized"
// @Failure      500   "Internal Server Error"
// @Router       /users/login [get]
func (hand *Handler) Login(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.Login"
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
		zap.String("route", "/users/login"),
	)

	rawUser := r.Header.Get("X-User")
	if rawUser == "" {
		span.SetStatus(codes.Error, "missing x-user")
		log.Info("unauthorized", zap.String("reason", "missing_x_user"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var user domain.User
	dec := json.NewDecoder(strings.NewReader(rawUser))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid x-user")
		log.Warn("unauthorized",
			zap.String("reason", "invalid_x_user"),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	span.SetAttributes(
		attribute.Int64("user.id", user.ID),
		attribute.String("user.role", user.Role),
	)

	out := hand.respConverter.ToAuthResponse(user)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(out); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "encode response failed")
		log.Error("encode response failed", zap.Error(err))
	}
}
