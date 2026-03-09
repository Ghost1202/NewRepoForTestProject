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

// Register godoc
// @Summary      Register a new user
// @Description  Creates a user. Default role: user.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      CreateUserReq  true  "User data"
// @Success      201    {object}  CreateUserResp
// @Failure      400   "Bad Request"
// @Failure      401   "Unauthorized"
// @Failure      409   "Conflict"
// @Failure      500   "Internal Server Error"
// @Router       /users/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.Register"
	tracer := otel.Tracer("/http/handler")

	ctx, span := tracer.Start(r.Context(), op,
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, h.log).With(
		zap.String("op", op),
		zap.String("route", "/users/register"),
	)

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Warn("body close error", zap.Error(err))
		}
	}()

	var reqBody CreateUserReq
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&reqBody); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode body failed")
		log.Error("failed to decode request body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reqBody.Email = strings.TrimSpace(reqBody.Email)
	reqBody.Login = strings.TrimSpace(reqBody.Login)
	reqBody.Password = strings.TrimSpace(reqBody.Password)

	if reqBody.Email == "" || reqBody.Password == "" || reqBody.Login == "" {
		span.SetStatus(codes.Error, "missing required fields")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("user.email", reqBody.Email),
		attribute.String("user.login", reqBody.Login),
	)

	userID, err := h.reqConverter.CreateUser(ctx, h.authService, reqBody)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			span.SetStatus(codes.Error, "user already exists")
			w.WriteHeader(http.StatusConflict)
			return
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "service error")
		log.Error("service error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.Int64("user.id", userID))

	out := h.respConverter.ToCreateUserResponse(userID)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(out); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "encode response failed")
		log.Error("encode response failed", zap.Error(err))
	}
}
