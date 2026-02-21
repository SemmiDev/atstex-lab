// Package middleware provides HTTP middleware for the application.
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

type contextKey string

const (
	// RequestIDKey is the context key for the request ID.
	RequestIDKey contextKey = "request_id"
	// TraceIDKey is the context key for the trace ID.
	TraceIDKey contextKey = "trace_id"
	// LoggerKey is the context key for the per-request logger.
	LoggerKey contextKey = "logger"
)

// generateTraceID produces a random 16-byte hex trace ID.
func generateTraceID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// RequestLogger is a slog-based HTTP request logger middleware.
// It replaces chi's default middleware.Logger with structured JSON output.
//
// Each request log line includes:
//   - request_id (from chi middleware.RequestID)
//   - trace_id   (generated per-request)
//   - method, path, remote_addr
//   - status, bytes_written, duration_ms
//
// It also injects a per-request *slog.Logger into the context so that
// downstream handlers can log with request_id and trace_id automatically.
func RequestLogger(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get request ID from chi (set by middleware.RequestID)
			requestID := chimw.GetReqID(r.Context())
			traceID := generateTraceID()

			// Build a per-request logger with request_id and trace_id baked in
			logger := base.With(
				slog.String("request_id", requestID),
				slog.String("trace_id", traceID),
			)

			// Inject IDs and logger into context
			ctx := r.Context()
			ctx = context.WithValue(ctx, RequestIDKey, requestID)
			ctx = context.WithValue(ctx, TraceIDKey, traceID)
			ctx = context.WithValue(ctx, LoggerKey, logger)

			// Wrap the response writer to capture status code and bytes
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)

			// Set response headers so clients can correlate
			ww.Header().Set("X-Request-ID", requestID)
			ww.Header().Set("X-Trace-ID", traceID)

			next.ServeHTTP(ww, r.WithContext(ctx))

			duration := time.Since(start)
			status := ww.Status()

			// Choose log level based on status code
			logFn := logger.Info
			if status >= 500 {
				logFn = logger.Error
			} else if status >= 400 {
				logFn = logger.Warn
			}

			logFn("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("query", r.URL.RawQuery),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.Int("status", status),
				slog.Int("bytes", ww.BytesWritten()),
				slog.Float64("duration_ms", float64(duration.Microseconds())/1000.0),
			)
		})
	}
}

// GetLogger retrieves the per-request *slog.Logger from the context.
// Falls back to slog.Default() if not found.
func GetLogger(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(LoggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// GetRequestID retrieves the request ID from context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// GetTraceID retrieves the trace ID from context.
func GetTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(TraceIDKey).(string); ok {
		return id
	}
	return ""
}
