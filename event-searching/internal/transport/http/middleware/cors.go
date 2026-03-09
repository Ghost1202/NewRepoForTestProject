package middleware

import (
	"net/http"

	config "github.com/turtlepavlo/event-searching/internal/transport/http"
)

func CORS(cfg config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.CORSEnabled {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", cfg.CORSAllowOrigins)
			w.Header().Set("Access-Control-Allow-Methods", cfg.CORSAllowMethods)
			w.Header().Set("Access-Control-Allow-Headers", cfg.CORSAllowHeaders)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
