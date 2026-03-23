package middleware

import (
	"context"
	"log/slog"
	"sync"
)

// WideEvent holds contextual structured data for the duration of a request.
type WideEvent struct {
	mu   sync.Mutex
	data map[string]any
}

// NewWideEvent creates a new WideEvent.
func NewWideEvent() *WideEvent {
	return &WideEvent{
		data: make(map[string]any),
	}
}

// Set adds or updates a key-value pair in the wide event.
func (w *WideEvent) Set(key string, value any) {
	if w == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	
	// Ensure we redact if necessary right away (optional) or at emission
	w.data[key] = value
}

// LogAttrs returns all accumulated data as a slice of slog.Any for the logger.
func (w *WideEvent) LogAttrs() []any {
	if w == nil {
		return nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	attrs := make([]any, 0, len(w.data))
	for k, v := range w.data {
		if r, ok := v.(Redactor); ok {
			attrs = append(attrs, slog.Any(k, r.Redact()))
		} else {
			attrs = append(attrs, slog.Any(k, v))
		}
	}
	return attrs
}

// AddEventData is a convenience helper to add data to the WideEvent in the context.
func AddEventData(ctx context.Context, key string, value any) {
	if w, ok := ctx.Value(WideEventKey).(*WideEvent); ok {
		w.Set(key, value)
	}
}
