package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/turtlepavlo/event-booking/docs"

	config "github.com/turtlepavlo/event-booking/internal/transport/http"
	mw "github.com/turtlepavlo/event-booking/internal/transport/http/middleware"
)

func NewRouter(hand *Handler, cfg config.Config, zapLog *zap.Logger, validator mw.TokenValidator) *mux.Router {
	rout := mux.NewRouter().StrictSlash(true)

	rout.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	rout.Use(mw.Recover(cfg, zapLog))
	rout.Use(mw.CORS(cfg))
	rout.Use(mw.AccessLog(cfg, zapLog))

	api := rout.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/tickets", hand.GetTickets).Methods(http.MethodGet, http.MethodPost)
	api.HandleFunc("/tickets/discounts", hand.GetEventDiscount).Methods(http.MethodGet, http.MethodPost)

	authMW := mw.Auth(validator, zapLog)

	bookings := api.PathPrefix("/bookings").Subrouter()
	bookings.Use(authMW)
	bookings.HandleFunc("", hand.CreateBooking).Methods(http.MethodPost)
	bookings.HandleFunc("/wallet", hand.OrderByWallet).Methods(http.MethodPost)
	bookings.HandleFunc("/refund", hand.RefundBooking).Methods(http.MethodPost)

	user := api.PathPrefix("/user").Subrouter()
	user.Use(authMW)
	user.HandleFunc("/tickets", hand.GetUserTickets).Methods(http.MethodGet, http.MethodPost)
	user.HandleFunc("/tickets/transfer", hand.TicketTransfer).Methods(http.MethodPost)
	user.HandleFunc("/waitlist", hand.AddToWaitlist).Methods(http.MethodPost)

	return rout
}
