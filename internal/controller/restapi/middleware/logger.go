package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

const reqIDKey contextKey = "request_id"

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logger(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)

		ctx := context.WithValue(r.Context(), reqIDKey, RequestIDKey)
		r = r.WithContext(ctx)
		w.Header().Set("X-Request-ID", string(RequestIDKey))

		next.ServeHTTP(rw, r)

		log.InfoContext(ctx, "http_request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rw.statusCode),
			slog.Duration("latency_ms", time.Since(start)),
			slog.String("request_id", string(RequestIDKey)),
		)
	})
}

func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(reqIDKey).(string); ok {
		return id
	}
	return ""
}
