package observability

import (
	"context"
	"log/slog"
	"strings"
)

type safeHandler struct {
	next  slog.Handler
	level slog.Handler
}

func (h *safeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.level.Enabled(ctx, level) && h.next.Enabled(ctx, level)
}

func (h *safeHandler) Handle(ctx context.Context, record slog.Record) error {
	message := record.Message
	switch message {
	case "HTTP request", "Starting server", "Shutting down server…", "Vue SPA frontend active at /":
	default:
		message = "Application event"
	}
	clean := slog.NewRecord(record.Time, record.Level, message, 0)
	record.Attrs(func(a slog.Attr) bool {
		if safe, ok := safeAttr(a); ok {
			clean.AddAttrs(safe)
		}
		return true
	})
	return h.next.Handle(ctx, clean)
}

func (h *safeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clean := make([]slog.Attr, 0, len(attrs))
	for _, a := range attrs {
		if safe, ok := safeAttr(a); ok {
			clean = append(clean, safe)
		}
	}
	return &safeHandler{next: h.next.WithAttrs(clean), level: h.level}
}

func (h *safeHandler) WithGroup(name string) slog.Handler {
	if name != "http" {
		return &safeHandler{next: h.next, level: h.level}
	}
	return &safeHandler{next: h.next.WithGroup(name), level: h.level}
}

func safeAttr(a slog.Attr) (slog.Attr, bool) {
	a.Value = a.Value.Resolve()
	switch a.Key {
	case "http.response.status_code":
		return a, a.Value.Kind() == slog.KindInt64
	case "http.request.duration":
		return a, a.Value.Kind() == slog.KindDuration
	case "http.request.method":
		if a.Value.Kind() != slog.KindString {
			return slog.Attr{}, false
		}
		switch a.Value.String() {
		case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "CONNECT", "TRACE", "_OTHER":
			return a, true
		}
	case "http.route":
		if a.Value.Kind() == slog.KindString {
			route := a.Value.String()
			if route == "unmatched" || (strings.HasPrefix(route, "/") && !strings.ContainsAny(route, "@?\r\n")) {
				return a, true
			}
		}
	case "error.type":
		if a.Value.Kind() == slog.KindString {
			switch a.Value.String() {
			case "server_error", "client_error", "request_error":
				return a, true
			}
		}
	case "http":
		if a.Value.Kind() == slog.KindGroup {
			attrs := make([]slog.Attr, 0, len(a.Value.Group()))
			for _, child := range a.Value.Group() {
				if safe, ok := safeAttr(child); ok {
					attrs = append(attrs, safe)
				}
			}
			return slog.Attr{Key: a.Key, Value: slog.GroupValue(attrs...)}, len(attrs) > 0
		}
	}
	return slog.Attr{}, false
}
