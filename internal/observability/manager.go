package observability

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Manager struct {
	cfg                  Config
	logger               *slog.Logger
	local                *slog.Logger
	traces               *sdktrace.TracerProvider
	logs                 *sdklog.LoggerProvider
	metrics              *sdkmetric.MeterProvider
	httpRequests         metric.Int64Counter
	httpDuration         metric.Float64Histogram
	dbOperations         metric.Int64Counter
	dbDuration           metric.Float64Histogram
	previousErrorHandler otel.ErrorHandler
	once                 sync.Once
	shutdownErr          error
	transport            *http.Transport
}

func New(ctx context.Context, cfg Config, version string, local *slog.Logger) (*Manager, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	m := &Manager{cfg: cfg, local: local, logger: local}
	if !cfg.Enabled {
		return m, nil
	}
	m.previousErrorHandler = otel.GetErrorHandler()
	otel.SetErrorHandler(&diagnostics{logger: local})
	m.transport = http.DefaultTransport.(*http.Transport).Clone()
	m.transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	m.transport.DisableKeepAlives = false
	client := &http.Client{
		Transport:     m.transport,
		Timeout:       cfg.ExportTimeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}
	res := resource.NewSchemaless(
		attribute.String("service.name", cfg.ServiceName),
		attribute.String("service.version", version),
		attribute.String("deployment.environment.name", cfg.Environment),
	)
	if cfg.AllowInsecureHTTP && m.usesHTTP() {
		local.Warn("Telemetry HTTP transport is unencrypted")
	}
	if cfg.TracesEnabled {
		exporter, err := otlptracehttp.New(ctx,
			otlptracehttp.WithEndpointURL(cfg.TracesEndpoint),
			otlptracehttp.WithHeaders(map[string]string{"Authorization": cfg.Authorization}),
			otlptracehttp.WithHTTPClient(client),
			otlptracehttp.WithTimeout(cfg.ExportTimeout),
			otlptracehttp.WithCompression(otlptracehttp.NoCompression),
			otlptracehttp.WithEncoding(otlptracehttp.EncodingProtobuf),
			otlptracehttp.WithRetry(otlptracehttp.RetryConfig{
				Enabled: true, InitialInterval: time.Second, MaxInterval: 2 * time.Second,
				MaxElapsedTime: cfg.ExportTimeout,
			}),
		)
		if err != nil {
			_ = m.Shutdown()
			return nil, fmt.Errorf("initialize telemetry trace exporter")
		}
		m.traces = sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.TraceSampleRatio))),
			sdktrace.WithRawSpanLimits(sdktrace.SpanLimits{
				AttributeValueLengthLimit: 512, AttributeCountLimit: 32,
				EventCountLimit: 16, LinkCountLimit: 16,
				AttributePerEventCountLimit: 16, AttributePerLinkCountLimit: 16,
			}),
			sdktrace.WithBatcher(exporter,
				sdktrace.WithMaxQueueSize(2048), sdktrace.WithMaxExportBatchSize(512),
				sdktrace.WithBatchTimeout(5*time.Second), sdktrace.WithExportTimeout(cfg.ExportTimeout),
			),
		)
	}
	if cfg.LogsEnabled {
		exporter, err := otlploghttp.New(ctx,
			otlploghttp.WithEndpointURL(cfg.LogsEndpoint),
			otlploghttp.WithHeaders(map[string]string{
				"Authorization": cfg.Authorization, "stream-name": cfg.LogsStream,
			}),
			otlploghttp.WithHTTPClient(client),
			otlploghttp.WithTimeout(cfg.ExportTimeout),
			otlploghttp.WithCompression(otlploghttp.NoCompression),
			otlploghttp.WithRetry(otlploghttp.RetryConfig{
				Enabled: true, InitialInterval: time.Second, MaxInterval: 2 * time.Second,
				MaxElapsedTime: cfg.ExportTimeout,
			}),
		)
		if err != nil {
			_ = m.Shutdown()
			return nil, fmt.Errorf("initialize telemetry log exporter")
		}
		m.logs = sdklog.NewLoggerProvider(
			sdklog.WithResource(res), sdklog.WithAttributeCountLimit(32),
			sdklog.WithAttributeValueLengthLimit(512),
			sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter,
				sdklog.WithMaxQueueSize(2048), sdklog.WithExportMaxBatchSize(512),
				sdklog.WithExportInterval(5*time.Second), sdklog.WithExportTimeout(cfg.ExportTimeout),
			)),
		)
		remote := &safeHandler{
			next:  otelslog.NewHandler("go-postfixadmin", otelslog.WithLoggerProvider(m.logs)),
			level: local.Handler(),
		}
		m.logger = slog.New(slog.NewMultiHandler(local.Handler(), remote))
	}
	if cfg.MetricsEnabled {
		if err := m.initializeMetrics(ctx, res, client); err != nil {
			_ = m.Shutdown()
			return nil, err
		}
	}
	return m, nil
}

func (m *Manager) Logger() *slog.Logger {
	return m.logger
}

func (m *Manager) usesHTTP() bool {
	return (m.cfg.TracesEnabled && strings.HasPrefix(m.cfg.TracesEndpoint, "http://")) ||
		(m.cfg.LogsEnabled && strings.HasPrefix(m.cfg.LogsEndpoint, "http://")) ||
		(m.cfg.MetricsEnabled && strings.HasPrefix(m.cfg.MetricsEndpoint, "http://"))
}

func (m *Manager) Shutdown() error {
	m.once.Do(func() {
		if !m.cfg.Enabled {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), m.cfg.ShutdownTimeout)
		defer cancel()
		var traceErr, logErr, metricErr error
		if m.traces != nil {
			traceErr = m.traces.Shutdown(ctx)
		}
		if m.logs != nil {
			logErr = m.logs.Shutdown(ctx)
		}
		if m.metrics != nil {
			metricErr = m.metrics.Shutdown(ctx)
		}
		m.shutdownErr = errors.Join(traceErr, logErr, metricErr)
		if m.shutdownErr != nil {
			m.local.Warn("Telemetry shutdown did not complete")
		}
		if m.transport != nil {
			m.transport.CloseIdleConnections()
		}
		otel.SetErrorHandler(m.previousErrorHandler)
	})
	return m.shutdownErr
}

type diagnostics struct {
	logger *slog.Logger
	mu     sync.Mutex
	last   time.Time
}

func (d *diagnostics) Handle(_ error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if time.Since(d.last) < 30*time.Second {
		return
	}
	d.last = time.Now()
	d.logger.Warn("Telemetry export degraded; check endpoint connectivity and ingestion settings")
}
