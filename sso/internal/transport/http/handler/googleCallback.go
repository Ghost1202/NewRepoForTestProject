package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/turtlepavlo/sso/internal/service"
	telemetry "github.com/turtlepavlo/sso/pkg/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// GoogleCallback godoc
// @Summary      Google OAuth callback (get JWT)
// @Description  Handles Google OAuth callback and returns JWT access token.
// @Description  If FRONTEND_URL is set, redirects to FRONTEND_URL/auth/callback?token=<JWT>.
// @Tags         users
// @Produce      json
// @Param        code  query     string  true  "Google auth code"
// @Success      200   {object}  LoginResp  "JSON response when FRONTEND_URL is empty"
// @Success      302   {string}  string     "Redirect to frontend callback with token when FRONTEND_URL is set"
// @Failure      400   "Bad Request"
// @Failure      401   "Unauthorized"
// @Failure      500   "Internal Server Error"
// @Router       /users/google/callback [get]
func (hand *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.GoogleCallback"
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
		zap.String("route", "/users/google/callback"),
	)

	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if code == "" {
		span.SetStatus(codes.Error, "missing code")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	session, err := hand.authService.LoginWithGoogle(ctx, code)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			span.SetStatus(codes.Error, "invalid credentials")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "login with google failed")
		log.Error("login with google failed", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	frontURL := strings.TrimSpace(hand.frontendURL)
	if frontURL != "" {
		token := session.AccessToken
		redirectURL := strings.TrimRight(frontURL, "/") + "/auth/callback?token=" + url.QueryEscape(token)
		http.Redirect(w, r, redirectURL, http.StatusFound)
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
