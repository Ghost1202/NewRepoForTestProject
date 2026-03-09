package middleware

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

func Tracing(serviceName string) func(http.Handler) http.Handler {
	if serviceName == "" {
		serviceName = "sso-service"
	}

	return otelmux.Middleware(
		serviceName,
		otelmux.WithSpanNameFormatter(func(route string, r *http.Request) string {
			return fmt.Sprintf("%s %s", r.Method, route)
		}),
	)
}
