package observability

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.opentelemetry.io/otel/codes"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestHTTPFinalStatus(t *testing.T) {
	cases := []struct {
		name    string
		handler echo.HandlerFunc
		status  int
	}{
		{"success", func(c *echo.Context) error { return c.NoContent(204) }, 204},
		{"error", func(c *echo.Context) error { return errors.New("password=secret") }, 500},
		{"http error", func(c *echo.Context) error { return echo.NewHTTPError(403, "secret") }, 403},
		{"panic", func(c *echo.Context) error { panic("private-user@example.test") }, 500},
	}
	for _, tc := range cases {
		for _, traces := range []bool{false, true} {
			name := tc.name + "/logs"
			if traces {
				name = tc.name + "/traces"
			}
			t.Run(name, func(t *testing.T) {
				var output bytes.Buffer
				recorder := tracetest.NewSpanRecorder()
				tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
				t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
				logger := slog.New(slog.NewJSONHandler(&output, nil))
				reader := sdkmetric.NewManualReader()
				mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
				t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })
				m := &Manager{cfg: Config{Enabled: true}, logger: logger, metrics: mp}
				if err := m.createInstruments(); err != nil {
					t.Fatal(err)
				}
				if traces {
					m.traces = tp
				}
				e := echo.New()
				m.Install(e)
				e.Use(middleware.Recover())
				e.GET("/mailboxes/:username", tc.handler)
				w := httptest.NewRecorder()
				e.ServeHTTP(w, httptest.NewRequest("GET", "/mailboxes/private-user@example.test?password=secret", nil))
				if w.Code != tc.status {
					t.Fatalf("status=%d", w.Code)
				}
				if strings.Count(output.String(), `"msg":"HTTP request"`) != 1 {
					t.Fatalf("duplicate or missing request log: %s", output.String())
				}
				if strings.Contains(output.String(), "secret") || strings.Contains(output.String(), "private-user") {
					t.Fatal("request data leaked")
				}
				var data metricdata.ResourceMetrics
				if err := reader.Collect(t.Context(), &data); err != nil {
					t.Fatal(err)
				}
				found := false
				for _, scope := range data.ScopeMetrics {
					for _, instrument := range scope.Metrics {
						if instrument.Name == "postfixadmin.http.requests" {
							points := instrument.Data.(metricdata.Sum[int64]).DataPoints
							if len(points) != 1 || points[0].Value != 1 {
								t.Fatal("request metric missing or duplicated")
							}
							value, ok := points[0].Attributes.Value("http.response.status_code")
							if !ok || value.AsInt64() != int64(tc.status) {
								t.Fatal("metric final status differs")
							}
							found = true
						}
					}
				}
				if !found {
					t.Fatal("no HTTP metrics")
				}
				if traces {
					ended := recorder.Ended()
					if len(ended) != 1 {
						t.Fatalf("ended spans=%d", len(ended))
					}
					span := ended[0]
					if span.Name() != "GET /mailboxes/:username" || span.SpanKind() != trace.SpanKindServer {
						t.Fatalf("span=%s", span.Name())
					}
					if tc.status >= 500 && span.Status().Code != codes.Error {
						t.Fatal("server error not recorded")
					}
				}
			})
		}
	}
}

func TestHTTPParentSampling(t *testing.T) {
	const parent = "00-11111111111111111111111111111111-2222222222222222-01"
	cases := []struct {
		name     string
		trust    bool
		ratio    float64
		exported bool
	}{
		{"untrusted zero", false, 0, false}, {"untrusted all", false, 1, true},
		{"trusted sampled", true, 0, true}, {"root zero", true, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := tracetest.NewSpanRecorder()
			tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(tc.ratio))), sdktrace.WithSpanProcessor(recorder))
			t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
			m := &Manager{cfg: Config{Enabled: true, TrustIncomingTraceContext: tc.trust}, logger: slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), traces: tp}
			e := echo.New()
			m.Install(e)
			var seen trace.SpanContext
			e.GET("/", func(c *echo.Context) error {
				seen = trace.SpanContextFromContext(c.Request().Context())
				return c.NoContent(200)
			})
			req := httptest.NewRequest("GET", "/", nil)
			if tc.name != "root zero" {
				req.Header.Set("traceparent", parent)
			}
			e.ServeHTTP(httptest.NewRecorder(), req)
			if seen.IsSampled() != tc.exported {
				t.Fatalf("sampling=%v", seen.IsSampled())
			}
			same := seen.TraceID().String() == "11111111111111111111111111111111"
			if same != (tc.trust && tc.name != "root zero") {
				t.Fatalf("trace=%s", seen.TraceID())
			}
		})
	}
}

func TestHTTPUnknownRoute(t *testing.T) {
	var output bytes.Buffer
	m := &Manager{cfg: Config{Enabled: true}, logger: slog.New(slog.NewJSONHandler(&output, nil))}
	e := echo.New()
	m.Install(e)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest("GET", "/private@example.test", nil))
	if w.Code != http.StatusNotFound || !strings.Contains(output.String(), "unmatched") || strings.Contains(output.String(), "private") {
		t.Fatalf("404: %s", output.String())
	}
}
