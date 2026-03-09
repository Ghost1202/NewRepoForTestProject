package handler

import (
	"net/http"

	telemetry "github.com/turtlepavlo/sso/pkg/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// GoogleLogin godoc
// @Summary      Google Login Redirect
// @Description  Redirects user to Google OAuth consent screen
// @Tags         users
// @Success      307  {string}  string "Redirect to Google"
// @Router       /users/google/login [get]
func (hand *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.GoogleLogin"
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
		zap.String("route", "/users/google/login"),
	)

	url := hand.authService.GetGoogleAuthURL(ctx)
	if url == "" {
		span.SetStatus(codes.Error, "empty google auth url")
		log.Error("google auth url is empty")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
