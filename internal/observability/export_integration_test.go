//go:build integration

package observability

import (
	"bytes"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	collectorlog "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	collectormetric "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/proto"
)

type payload struct {
	path, authorization, stream, contentType string
	data                                     []byte
}

func TestOTLPExport(t *testing.T) {
	var mu sync.Mutex
	received := []payload{}
	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		mu.Lock()
		received = append(received, payload{path: r.URL.Path, authorization: r.Header.Get("Authorization"), stream: r.Header.Get("stream-name"), contentType: r.Header.Get("Content-Type"), data: data})
		mu.Unlock()
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(200)
	}))
	defer collector.Close()
	v := enabledConfig(t)
	v.Set("observability.allow_insecure_http", true)
	v.Set("observability.traces_endpoint", collector.URL+"/custom/org/v1/traces")
	v.Set("observability.logs_endpoint", collector.URL+"/custom/org/v1/logs")
	v.Set("observability.trace_sample_ratio", 1)
	v.Set("observability.metrics_enabled", true)
	v.Set("observability.metrics_endpoint", collector.URL+"/custom/org/v1/metrics")
	for key, value := range map[string]string{
		"OTEL_EXPORTER_OTLP_ENDPOINT":                              "http://bad.invalid/wrong",
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT":                         "http://bad.invalid/wrong-logs",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT":                       "http://bad.invalid/wrong-traces",
		"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT":                      "http://bad.invalid/wrong-metrics",
		"OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE":        "delta",
		"OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION": "base2_exponential_bucket_histogram",
		"OTEL_EXPORTER_OTLP_HEADERS":                               "Authorization=private-token",
		"OTEL_EXPORTER_OTLP_INSECURE":                              "true",
		"OTEL_BSP_MAX_QUEUE_SIZE":                                  "1", "OTEL_BLRP_MAX_QUEUE_SIZE": "1",
		"OTEL_TRACES_SAMPLER": "always_off", "OTEL_SDK_DISABLED": "true",
	} {
		t.Setenv(key, value)
	}
	cfg, err := LoadConfig(v)
	if err != nil {
		t.Fatal(err)
	}
	var local bytes.Buffer
	m, err := New(t.Context(), cfg, "integration", slog.New(slog.NewJSONHandler(&local, nil)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = m.Shutdown() })
	e := echo.New()
	m.Install(e)
	e.GET("/mailboxes/:username", func(c *echo.Context) error {
		m.Logger().InfoContext(c.Request().Context(), "secret password=private-token", "password", "private-token", "error", "user@example.test", "dsn", "secret", "client.address", "127.0.0.1")
		m.Logger().With("password", "private-token").WithGroup("secret-group").InfoContext(c.Request().Context(), "secret-message")
		m.Logger().DebugContext(c.Request().Context(), "debug-secret")
		return c.NoContent(204)
	})
	req := httptest.NewRequest("GET", "/mailboxes/user@example.test?token=private-token", nil)
	req.Header.Set("Authorization", "Bearer private-token")
	e.ServeHTTP(httptest.NewRecorder(), req)
	if err := m.Shutdown(); err != nil {
		t.Fatal(err)
	}
	if err := m.Shutdown(); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	var traceID, spanID string
	requestLogs := 0
	metricFound := false
	for _, p := range received {
		if p.authorization != "Basic test" || p.contentType != "application/x-protobuf" {
			t.Fatalf("OTLP headers = %+v", p)
		}
		if bytes.Contains(p.data, []byte("private-token")) || bytes.Contains(p.data, []byte("user@example.test")) || bytes.Contains(p.data, []byte("secret-group")) || bytes.Contains(p.data, []byte("debug-secret")) {
			t.Fatalf("sensitive data exported at %s: %q", p.path, string(p.data))
		}
		if bytes.Contains(p.data, []byte("wrong")) {
			t.Fatal("SDK env overrode explicit configuration")
		}
		switch p.path {
		case "/custom/org/v1/traces":
			body := new(collectortrace.ExportTraceServiceRequest)
			if err := proto.Unmarshal(p.data, body); err != nil {
				t.Fatal(err)
			}
			if len(body.ResourceSpans) != 1 {
				t.Fatal("missing trace resource")
			}
			for _, scope := range body.ResourceSpans[0].ScopeSpans {
				for _, span := range scope.Spans {
					if span.Name != "GET /mailboxes/:username" {
						t.Fatalf("span=%s", span.Name)
					}
					traceID = hex.EncodeToString(span.TraceId)
					spanID = hex.EncodeToString(span.SpanId)
				}
			}
		case "/custom/org/v1/metrics":
			body := new(collectormetric.ExportMetricsServiceRequest)
			if err := proto.Unmarshal(p.data, body); err != nil {
				t.Fatal(err)
			}
			for _, resource := range body.ResourceMetrics {
				for _, scope := range resource.ScopeMetrics {
					for _, instrument := range scope.Metrics {
						if instrument.Name == "http.server.request.duration" {
							points := instrument.GetHistogram().DataPoints
							if len(points) != 1 || len(points[0].ExplicitBounds) != 14 || points[0].ExplicitBounds[0] != 0.005 || points[0].ExplicitBounds[13] != 10 {
								t.Fatal("HTTP histogram bucket boundaries differ")
							}
						}
						if instrument.Name == "postfixadmin.http.requests" && len(instrument.GetSum().DataPoints) == 1 && instrument.GetSum().DataPoints[0].GetAsInt() == 1 {
							metricFound = true
						}
					}
				}
			}
		case "/custom/org/v1/logs":
			if p.stream != "postfixadmin" {
				t.Fatal("wrong logs stream")
			}
			body := new(collectorlog.ExportLogsServiceRequest)
			if err := proto.Unmarshal(p.data, body); err != nil {
				t.Fatal(err)
			}
			for _, resource := range body.ResourceLogs {
				for _, scope := range resource.ScopeLogs {
					for _, record := range scope.LogRecords {
						if record.Body.GetStringValue() == "HTTP request" {
							requestLogs++
							if len(record.TraceId) != 16 || len(record.SpanId) != 8 {
								t.Fatal("missing log correlation")
							}
						}
					}
				}
			}
		default:
			t.Fatalf("unexpected path %s", p.path)
		}
	}
	if !metricFound {
		t.Fatal("OTLP metric export missing")
	}
	if traceID == "" || spanID == "" || requestLogs != 1 {
		t.Fatalf("trace=%s, requests=%d", traceID, requestLogs)
	}
	for _, p := range received {
		if strings.HasSuffix(p.path, "/logs") {
			body := new(collectorlog.ExportLogsServiceRequest)
			_ = proto.Unmarshal(p.data, body)
			for _, resource := range body.ResourceLogs {
				for _, scope := range resource.ScopeLogs {
					for _, record := range scope.LogRecords {
						if record.Body.GetStringValue() == "HTTP request" && (hex.EncodeToString(record.TraceId) != traceID || hex.EncodeToString(record.SpanId) != spanID) {
							t.Fatal("log/trace IDs differ")
						}
					}
				}
			}
		}
	}
	if !strings.Contains(local.String(), "private-token") {
		t.Fatal("local behavior changed")
	}
}

func TestExportOutage(t *testing.T) {
	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(503) }))
	defer collector.Close()
	v := enabledConfig(t)
	v.Set("observability.traces_enabled", false)
	v.Set("observability.allow_insecure_http", true)
	v.Set("observability.logs_endpoint", collector.URL+"/logs")
	v.Set("observability.metrics_enabled", true)
	v.Set("observability.metrics_endpoint", collector.URL+"/metrics")
	v.Set("observability.metrics_export_interval", "10ms")
	v.Set("observability.export_timeout", "100ms")
	v.Set("observability.shutdown_timeout", "200ms")
	cfg, err := LoadConfig(v)
	if err != nil {
		t.Fatal(err)
	}
	var local bytes.Buffer
	m, err := New(t.Context(), cfg, "test", slog.New(slog.NewTextHandler(&local, nil)))
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	for range 10000 {
		m.Logger().Info("outage-secret")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("logging blocked: %s", elapsed)
	}
	start = time.Now()
	_ = m.Shutdown()
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("shutdown exceeded budget: %s", elapsed)
	}
	if strings.Count(local.String(), "Telemetry export degraded") > 1 {
		t.Fatal("export diagnostic flood")
	}
}
