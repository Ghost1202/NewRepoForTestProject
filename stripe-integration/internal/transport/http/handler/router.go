package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	cfg "github.com/turtlepavlo/stripe_integration/internal/transport/http"
	"github.com/turtlepavlo/stripe_integration/internal/transport/http/middleware"
)

func NewRouter(h *Handler, cfg cfg.Config, log *zap.Logger) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	r.Use(middleware.Recover(cfg, log))
	r.Use(middleware.AccessLog(cfg, log))

	r.HandleFunc("/webhook", h.HandleStripeWebhook).Methods(http.MethodPost)

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("OK")); err != nil {
			_ = err
		}
	}).Methods(http.MethodGet)
	return r
}
