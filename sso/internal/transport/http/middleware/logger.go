package middleware

import (
	"bufio"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/turtlepavlo/sso/internal/domain"
)

type statusCapturingResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (w *statusCapturingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusCapturingResponseWriter) Write(p []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytesWritten += n
	return n, err
}

func (w *statusCapturingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return h.Hijack()
}

func (w *statusCapturingResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *statusCapturingResponseWriter) Push(target string, opts *http.PushOptions) error {
	p, ok := w.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}
	return p.Push(target, opts)
}

func (w *statusCapturingResponseWriter) ReadFrom(r io.Reader) (int64, error) {
	rf, ok := w.ResponseWriter.(io.ReaderFrom)
	if !ok {
		return 0, http.ErrNotSupported
	}
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	n, err := rf.ReadFrom(r)
	w.bytesWritten += int(n)
	return n, err
}

func MwLogger(cfg Config, zapLog *zap.Logger) func(http.Handler) http.Handler {
	if !cfg.LoggerEnabled {
		return func(next http.Handler) http.Handler { return next }
	}
	if zapLog == nil {
		zapLog = zap.NewNop()
	}

	httpLog := zapLog.With(
		zap.String("layer", "transport"),
		zap.String("component", "http"),
		zap.String("op", "middleware.MwLogger"),
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusCapturingResponseWriter{ResponseWriter: w}

			next.ServeHTTP(sw, r)

			status := sw.statusCode
			if status == 0 {
				status = http.StatusOK
			}

			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("query", r.URL.RawQuery),
				zap.Int("status", status),
				zap.Int("bytes", sw.bytesWritten),
				zap.Duration("duration", time.Since(start)),
				zap.String("user_agent", r.UserAgent()),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("host", r.Host),
			}

			if sc := trace.SpanContextFromContext(r.Context()); sc.IsValid() {
				fields = append(fields,
					zap.String("trace_id", sc.TraceID().String()),
					zap.String("span_id", sc.SpanID().String()),
				)
			}

			if rawUser := r.Header.Get(userHeaderKey); rawUser != "" {
				var u domain.User
				if err := json.Unmarshal([]byte(rawUser), &u); err == nil {
					fields = append(fields,
						zap.Int64("user_id", u.ID),
						zap.String("role", u.Role),
					)
				}
			}

			switch {
			case status >= 500:
				httpLog.Error("request", fields...)
			case status >= 400:
				httpLog.Warn("request", fields...)
			default:
				httpLog.Info("request", fields...)
			}
		})
	}
}
