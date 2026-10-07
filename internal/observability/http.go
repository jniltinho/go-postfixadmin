package observability

import (
	"log/slog"
	"net/http"
	"time"

	"context"
	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const requestStateKey = "go-postfixadmin.observability.request"

type requestState struct {
	started  time.Time
	span     trace.Span
	finished bool
}

func (m *Manager) Install(e *echo.Echo) {
	if !m.cfg.Enabled {
		return
	}
	original := e.HTTPErrorHandler
	e.HTTPErrorHandler = func(c *echo.Context, err error) {
		original(c, err)
		m.finishRequest(c, err)
	}
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			state := &requestState{started: time.Now()}
			c.Set(requestStateKey, state)
			c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), databaseRequestKey{}, true)))
			if m.traces != nil {
				ctx := c.Request().Context()
				if m.cfg.TrustIncomingTraceContext {
					ctx = (propagation.TraceContext{}).Extract(ctx, propagation.HeaderCarrier(c.Request().Header))
				}
				options := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindServer)}
				if !m.cfg.TrustIncomingTraceContext {
					options = append(options, trace.WithNewRoot())
				}
				ctx, state.span = m.traces.Tracer("go-postfixadmin/http").Start(
					ctx, requestMethod(c.Request().Method)+" "+requestRoute(c), options...,
				)
				c.SetRequest(c.Request().WithContext(ctx))
			}
			err := next(c)
			if err == nil {
				m.finishRequest(c, nil)
			}
			return err
		}
	})
}

func (m *Manager) finishRequest(c *echo.Context, err error) {
	state, ok := c.Get(requestStateKey).(*requestState)
	if !ok || state.finished {
		return
	}
	state.finished = true
	_, status := echo.ResolveResponseStatus(c.Response(), nil)
	method, route := requestMethod(c.Request().Method), requestRoute(c)
	duration := time.Since(state.started)
	attrs := []slog.Attr{
		slog.String("http.request.method", method), slog.String("http.route", route),
		slog.Int("http.response.status_code", status), slog.Duration("http.request.duration", duration),
	}
	if state.span != nil {
		state.span.SetName(method + " " + route)
		state.span.SetAttributes(
			attribute.String("http.request.method", method), attribute.String("http.route", route),
			attribute.Int("http.response.status_code", status),
		)
	}
	if m.metrics != nil {
		options := metric.WithAttributes(attribute.String("http.request.method", method),
			attribute.String("http.route", route), attribute.Int("http.response.status_code", status))
		m.httpRequests.Add(c.Request().Context(), 1, options)
		m.httpDuration.Record(c.Request().Context(), duration.Seconds(), options)
	}
	level := slog.LevelInfo
	if status >= http.StatusInternalServerError {
		level = slog.LevelError
		attrs = append(attrs, slog.String("error.type", "server_error"))
		if state.span != nil {
			state.span.SetStatus(codes.Error, "server_error")
			state.span.AddEvent("request.error", trace.WithAttributes(attribute.String("error.type", "server_error")))
		}
	} else if err != nil {
		attrs = append(attrs, slog.String("error.type", "request_error"))
	}
	m.logger.LogAttrs(c.Request().Context(), level, "HTTP request", attrs...)
	if state.span != nil {
		state.span.End()
	}
}

func requestMethod(method string) string {
	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "CONNECT", "TRACE":
		return method
	}
	return "_OTHER"
}

func requestRoute(c *echo.Context) string {
	if c.Path() == "" {
		return "unmatched"
	}
	return c.Path()
}
