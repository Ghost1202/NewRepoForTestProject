package middleware

import "net/http"

func CORS(cfg Config) func(http.Handler) http.Handler {
	if !cfg.CORSEnabled {
		return func(next http.Handler) http.Handler { return next }
	}

	allowOrigins := cfg.CORSAllowOrigins
	allowMethods := cfg.CORSAllowMethods
	allowHeaders := cfg.CORSAllowHeaders
	allowCredentials := cfg.CORSAllowCredentials

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			headers := responseWriter.Header()
			if allowCredentials && allowOrigins == "*" {
				if origin := request.Header.Get("Origin"); origin != "" {
					headers.Set("Access-Control-Allow-Origin", origin)
				} else {
					headers.Set("Access-Control-Allow-Origin", allowOrigins)
				}
			} else {
				headers.Set("Access-Control-Allow-Origin", allowOrigins)
			}
			headers.Set("Access-Control-Allow-Methods", allowMethods)
			headers.Set("Access-Control-Allow-Headers", allowHeaders)
			if allowCredentials {
				headers.Set("Access-Control-Allow-Credentials", "true")
			}
			headers.Set("Vary", "Origin")

			if request.Method == http.MethodOptions {
				responseWriter.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(responseWriter, request)
		})
	}
}
