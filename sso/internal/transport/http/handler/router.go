package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.uber.org/zap"

	_ "github.com/turtlepavlo/sso/internal/docs"
	mw "github.com/turtlepavlo/sso/internal/transport/http/middleware"
)

func NewRouter(hand *Handler, cfg mw.Config, zapLog *zap.Logger, jwtSecret []byte, serviceName string) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	r.Use(otelmux.Middleware(serviceName))

	r.Use(mw.Recover(cfg, zapLog))
	r.Use(mw.MwLogger(cfg, zapLog))
	r.Use(mw.CORS(cfg))
	r.Use(mw.LeakyBucketMiddleware(cfg, zapLog))

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	r.HandleFunc("/users/register", hand.Register).Methods(http.MethodPost)
	r.HandleFunc("/users/auth", hand.Auth).Methods(http.MethodPost)

	r.HandleFunc("/users/google/login", hand.GoogleLogin).Methods(http.MethodGet)
	r.HandleFunc("/users/google/callback", hand.GoogleCallback).Methods(http.MethodGet)

	r.HandleFunc("/users/auth/phone", hand.AuthByPhone).Methods(http.MethodPost)

	r.HandleFunc("/users/otp/phone/send", hand.SendOTPByPhone).Methods(http.MethodPost)
	r.HandleFunc("/users/otp/phone/confirm", hand.ConfirmOTPByPhone).Methods(http.MethodPost)

	r.HandleFunc("/users/reset/phone", hand.ResetPhone).Methods(http.MethodPost)
	r.HandleFunc("/users/reset/phone/confirm", hand.ResetPhoneConfirm).Methods(http.MethodPost)
	r.HandleFunc("/users/reset/email", hand.ResetEmail).Methods(http.MethodPost)
	r.HandleFunc("/users/reset/email/confirm", hand.ResetEmailConfirm).Methods(http.MethodPost)

	protected := r.PathPrefix("/users").Subrouter()
	protected.Use(mw.Auth(jwtSecret, zapLog))
	protected.HandleFunc("/login", hand.Login).Methods(http.MethodGet)

	return r
}
