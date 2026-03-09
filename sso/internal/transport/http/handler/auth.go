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

// Auth godoc
// @Summary      Authorization by login/email (get JWT)
// @Description  Validates identifier/password and returns JWT access token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      LoginReq  true  "Credentials"
// @Success      200   {object}  LoginResp
// @Failure      400   "Bad Request"
// @Failure      401   "Unauthorized"
// @Failure      500   "Internal Server Error"
// @Router       /users/auth [post]
func (hand *Handler) Auth(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.Auth"
	tracer := otel.Tracer("http/handler")

	ctx, span := tracer.Start(r.Context(), op,
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, hand.log).With(
		zap.String("op", op),
		zap.String("route", "/users/auth"),
	)

	var reqBody LoginReq
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&reqBody); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode body failed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reqBody.Identifier = strings.TrimSpace(reqBody.Identifier)
	reqBody.Password = strings.TrimSpace(reqBody.Password)
	span.SetAttributes(attribute.Bool("auth.is_email", reqBody.IsEmail))

	session, err := hand.reqConverter.Login(ctx, hand.authService, reqBody)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			span.SetStatus(codes.Error, "invalid credentials")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "service login failed")
		log.Error("service login failed", zap.Error(err))
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
