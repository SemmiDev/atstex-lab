// Package middleware provides HTTP middleware for the application.
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/semmidev/atstex-lab/internal/apperrors"
)

type contextKey string

const (
	// RequestIDKey is the context key for the request ID.
	RequestIDKey contextKey = "request_id"
	// TraceIDKey is the context key for the trace ID.
	TraceIDKey contextKey = "trace_id"
	// LoggerKey is the context key for the per-request logger.
	LoggerKey contextKey = "logger"
	// WideEventKey is the context key for the per-request wide event.
	WideEventKey contextKey = "wide_event"
)

// Respond is a helper for successful JSON responses.
func Respond(w http.ResponseWriter, r *http.Request, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		if err := json.NewEncoder(w).Encode(v); err != nil {
			AddEventData(r.Context(), "json_encode_error", err.Error())
		}
	}
}

// RespondError is the ONLY approved path for writing error responses.
// It translates any error into an RFC 7807 Problem Details object.
func RespondError(w http.ResponseWriter, r *http.Request, err error) {
	prob := apperrors.Translate(err)

	// Append error details to the wide event instead of immediately logging
	var safe *apperrors.SafeError
	if errors.As(err, &safe) {
		AddEventData(r.Context(), "app_error", map[string]any{
			"details": safe.LogString(),
			"code":    safe.Code,
		})
	} else {
		AddEventData(r.Context(), "unhandled_error", err.Error())
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(prob.Status)
	if encodeErr := json.NewEncoder(w).Encode(prob); encodeErr != nil {
		AddEventData(r.Context(), "problem_encode_error", encodeErr.Error())
	}
}

// generateTraceID produces a random 16-byte hex trace ID.
func generateTraceID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
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

			// Initialize wide event
			we := NewWideEvent()

			safeHeaders := r.Header.Clone()
			safeHeaders.Del("Authorization")
			safeHeaders.Del("Cookie")
			we.Set("http_headers", safeHeaders)

			// Inject IDs, logger, and wide event into context
			ctx := r.Context()
			ctx = context.WithValue(ctx, RequestIDKey, requestID)
			ctx = context.WithValue(ctx, TraceIDKey, traceID)
			ctx = context.WithValue(ctx, LoggerKey, logger)
			ctx = context.WithValue(ctx, WideEventKey, we)

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

			// Add final metrics to wide event
			we.Set("http_request", map[string]any{
				"method":      r.Method,
				"path":        r.URL.Path,
				"query":       r.URL.RawQuery,
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
			})
			we.Set("status", status)
			we.Set("bytes", ww.BytesWritten())
			we.Set("duration_ms", float64(duration.Microseconds())/1000.0)

			// Emit a single comprehensive log line
			logFn("http_request_completed", we.LogAttrs()...)
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
