package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	_ "github.com/turtlepavlo/event-searching/docs" // swagger docs

	config "github.com/turtlepavlo/event-searching/internal/transport/http"
	mw "github.com/turtlepavlo/event-searching/internal/transport/http/middleware"
)

func NewRouter(hand *Handler, cfg config.Config, zapLog *zap.Logger, validator mw.TokenValidator) *mux.Router {
	rout := mux.NewRouter().StrictSlash(true)

	rout.Use(mw.Recover(cfg, zapLog))
	rout.Use(mw.CORS(cfg))
	rout.Use(mw.AccessLog(cfg, zapLog))

	rout.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	api := rout.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/events", hand.GetEvents).Methods(http.MethodGet)
	api.HandleFunc("/events/search", hand.GetEventsFilter).Methods(http.MethodGet)

	api.HandleFunc("/comments/list", hand.GetComments).Methods(http.MethodPost)
	api.HandleFunc("/comments/rating", hand.GetRating).Methods(http.MethodPost)

	private := api.NewRoute().Subrouter()
	private.Use(mw.Auth(validator, zapLog))
	private.HandleFunc("/comments", hand.SetComment).Methods(http.MethodPost)

	return rout
}
