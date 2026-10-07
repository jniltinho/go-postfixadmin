package observability

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/codes"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type telemetryDriver struct{}
type telemetryConn struct{}
type telemetryRows struct{ done bool }

func (telemetryDriver) Open(string) (driver.Conn, error)  { return telemetryConn{}, nil }
func (telemetryConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unsupported") }
func (telemetryConn) Close() error                        { return nil }
func (telemetryConn) Begin() (driver.Tx, error)           { return nil, errors.New("unsupported") }
func (telemetryConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	if strings.Contains(query, "broken") {
		return nil, errors.New("SQL password=secret user@example.test")
	}
	return driver.RowsAffected(1), nil
}
func (telemetryConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	return &telemetryRows{done: strings.Contains(query, "not_found")}, nil
}
func (*telemetryRows) Columns() []string { return []string{"id"} }
func (*telemetryRows) Close() error      { return nil }
func (r *telemetryRows) Next(values []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	values[0] = int64(1)
	return nil
}
func init() { sql.Register("postfixadmin-telemetry-test", telemetryDriver{}) }

func testDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	pool, err := sql.Open("postfixadmin-telemetry-test", "password=secret")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: pool, SkipInitializeWithVersion: true}),
		&gorm.Config{SkipDefaultTransaction: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestDatabaseSignals(t *testing.T) {
	for _, dbTraces := range []bool{false, true} {
		for _, metrics := range []bool{false, true} {
			t.Run(strings.Join([]string{map[bool]string{false: "no-traces", true: "traces"}[dbTraces], map[bool]string{false: "no-metrics", true: "metrics"}[metrics]}, "/"), func(t *testing.T) {
				recorder := tracetest.NewSpanRecorder()
				tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
				t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
				reader := sdkmetric.NewManualReader()
				mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
				t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })
				m := &Manager{cfg: Config{Enabled: true, DatabaseTracesEnabled: dbTraces, MetricsEnabled: metrics}, traces: tp}
				if metrics {
					m.metrics = mp
					if err := m.createInstruments(); err != nil {
						t.Fatal(err)
					}
				}
				db := testDatabase(t)
				if err := m.InstallDatabase(db); err != nil {
					t.Fatal(err)
				}
				ctx, parent := tp.Tracer("test").Start(context.WithValue(context.Background(), databaseRequestKey{}, true), "HTTP")
				requestDB := db.WithContext(ctx)
				var rows []struct{ ID int }
				if err := requestDB.Raw("SELECT id FROM private_table WHERE email = ?", "user@example.test").Scan(&rows).Error; err != nil {
					t.Fatal(err)
				}
				if err := requestDB.Exec("broken password=secret").Error; err == nil {
					t.Fatal("expected database error")
				}
				if err := requestDB.Table("not_found").First(&struct{ ID int }{}).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
					t.Fatalf("not found error=%v", err)
				}
				if err := db.Exec("UPDATE private_table SET active=1").Error; err != nil {
					t.Fatal(err)
				}
				parent.End()
				if db.Statement.Context.Value(databaseRequestKey{}) != nil {
					t.Fatal("shared context mutated")
				}
				spans := recorder.Ended()
				expected := 1
				if dbTraces {
					expected = 4
				}
				if len(spans) != expected {
					t.Fatalf("spans=%d want%d", len(spans), expected)
				}
				for _, span := range spans {
					if span.Name() == "HTTP" {
						continue
					}
					if span.Parent().SpanID() != parent.SpanContext().SpanID() || span.SpanContext().TraceID() != parent.SpanContext().TraceID() || span.SpanKind() != trace.SpanKindClient {
						t.Fatal("database span lost HTTP parent")
					}
					for _, attr := range span.Attributes() {
						if strings.Contains(attr.Value.Emit(), "secret") || strings.Contains(attr.Value.Emit(), "example.test") || strings.Contains(string(attr.Key), "query") {
							t.Fatal("SQL or secrets exported")
						}
					}
					if span.Name() == "DB SELECT" && span.Status().Code == codes.Error {
						t.Fatal("expected not-found recorded as error")
					}
					if span.Name() == "DB EXEC" && span.Status().Code != codes.Error {
						t.Fatal("database error status absent")
					}
				}
				if metrics {
					var data metricdata.ResourceMetrics
					if err := reader.Collect(t.Context(), &data); err != nil {
						t.Fatal(err)
					}
					found := false
					for _, scope := range data.ScopeMetrics {
						for _, instrument := range scope.Metrics {
							if instrument.Name == "postfixadmin.db.operations" {
								sum := instrument.Data.(metricdata.Sum[int64])
								var total int64
								for _, point := range sum.DataPoints {
									total += point.Value
								}
								if total != 3 {
									t.Fatalf("operations=%d", total)
								}
								found = true
							}
						}
					}
					if !found {
						t.Fatal("no database metrics")
					}
				}
			})
		}
	}
}
