package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/turtlepavlo/event-service/docs"
	transport "github.com/turtlepavlo/event-service/internal/transport/http"
	"github.com/turtlepavlo/event-service/internal/transport/http/middleware"
)

func NewRouter(hand *Handler, cfg transport.Config, zapLog *zap.Logger, authValidator middleware.TokenValidator) *mux.Router {
	rout := mux.NewRouter().StrictSlash(true)

	rout.Use(middleware.Recover(cfg, zapLog))
	rout.Use(middleware.CORS(cfg))
	rout.Use(middleware.AccessLog(cfg, zapLog))

	rout.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	api := rout.PathPrefix("/api/v1").Subrouter()

	api.Use(middleware.Auth(authValidator, zapLog))

	api.HandleFunc("/events", hand.CreateEvent).Methods(http.MethodPost)
	api.HandleFunc("/events", hand.UpdateEvent).Methods(http.MethodPut)

	api.HandleFunc("/promos/promo", hand.AddPromo).Methods(http.MethodPost)
	api.HandleFunc("/promos/early", hand.AddEarly).Methods(http.MethodPost)
	api.HandleFunc("/promos/bundle", hand.AddBundle).Methods(http.MethodPost)

	api.HandleFunc("/promos/promo/search", hand.GetPromo).Methods(http.MethodPost)
	api.HandleFunc("/promos/early/search", hand.GetEarly).Methods(http.MethodPost)
	api.HandleFunc("/promos/bundle/search", hand.GetBundle).Methods(http.MethodPost)

	return rout
}
