package http

import (
	"log/slog"
	"net/http"
	"time"

	"nba-games-service/internal/logging"
)

// LoggingMiddleware wraps the handler with request logging and request ID support.
func LoggingMiddleware(baseLogger *slog.Logger, next http.Handler) http.Handler {
	if baseLogger == nil {
		baseLogger = slog.Default()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = generateRequestID()
		}
		w.Header().Set("X-Request-ID", reqID)

		clientIP := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			clientIP = forwarded
		}

		logger := baseLogger.With(
			slog.String("request_id", reqID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("query", r.URL.RawQuery),
			slog.String("client_ip", clientIP),
		)

		ctx := logging.WithLogger(r.Context(), logger)
		ctx = withRequestID(ctx, reqID)
		r = r.WithContext(ctx)
		ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(ww, r)

		logger.Info("request complete",
			slog.Int("status", ww.status),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
	})
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

type responseWriter struct {
	http.ResponseWriter
	status int
}
