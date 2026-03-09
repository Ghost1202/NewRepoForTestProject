package middleware

import (
	"net/http"

	transport "github.com/turtlepavlo/event-service/internal/transport/http"
)

func CORS(cfg transport.Config) func(http.Handler) http.Handler {
	if !cfg.CORSEnabled {
		return func(next http.Handler) http.Handler { return next }
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.Header().Set("Access-Control-Allow-Origin", cfg.CORSAllowOrigins)
			responseWriter.Header().Set("Access-Control-Allow-Methods", cfg.CORSAllowMethods)
			responseWriter.Header().Set("Access-Control-Allow-Headers", cfg.CORSAllowHeaders)

			if request.Method == http.MethodOptions {
				responseWriter.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(responseWriter, request)
		})
	}
}
