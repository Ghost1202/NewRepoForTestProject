package middleware

import (
	"net/http"

	config "github.com/turtlepavlo/event-booking/internal/transport/http"
)

func CORS(config config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !config.CORSEnabled {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", config.CORSAllowOrigins)
			w.Header().Set("Access-Control-Allow-Methods", config.CORSAllowMethods)
			w.Header().Set("Access-Control-Allow-Headers", config.CORSAllowHeaders)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
