package observability

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func (m *Manager) initializeMetrics(ctx context.Context, res *resource.Resource, client *http.Client) error {
	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpointURL(m.cfg.MetricsEndpoint),
		otlpmetrichttp.WithHeaders(map[string]string{"Authorization": m.cfg.Authorization}),
		otlpmetrichttp.WithHTTPClient(client),
		otlpmetrichttp.WithTimeout(m.cfg.ExportTimeout),
		otlpmetrichttp.WithCompression(otlpmetrichttp.NoCompression),
		otlpmetrichttp.WithTemporalitySelector(sdkmetric.DefaultTemporalitySelector),
		otlpmetrichttp.WithAggregationSelector(sdkmetric.DefaultAggregationSelector),
		otlpmetrichttp.WithRetry(otlpmetrichttp.RetryConfig{Enabled: true,
			InitialInterval: time.Second, MaxInterval: 2 * time.Second, MaxElapsedTime: m.cfg.ExportTimeout}),
	)
	if err != nil {
		return fmt.Errorf("initialize telemetry metric exporter")
	}
	m.metrics = sdkmetric.NewMeterProvider(sdkmetric.WithResource(res), sdkmetric.WithCardinalityLimit(2000),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(m.cfg.MetricsExportInterval), sdkmetric.WithTimeout(m.cfg.ExportTimeout))))
	return m.createInstruments()
}

func (m *Manager) createInstruments() error {
	meter := m.metrics.Meter("go-postfixadmin")
	var err error
	m.httpRequests, err = meter.Int64Counter("postfixadmin.http.requests", metric.WithUnit("{request}"))
	if err != nil {
		return err
	}
	m.httpDuration, err = meter.Float64Histogram("http.server.request.duration", metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(.005, .01, .025, .05, .075, .1, .25, .5, .75, 1, 2.5, 5, 7.5, 10))
	if err != nil {
		return err
	}
	m.dbOperations, err = meter.Int64Counter("postfixadmin.db.operations", metric.WithUnit("{operation}"))
	if err != nil {
		return err
	}
	m.dbDuration, err = meter.Float64Histogram("db.client.operation.duration", metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10))
	return err
}
