package observability

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type databaseRequestKey struct{}
type databaseStateKey struct{}
type databaseState struct {
	parent  context.Context
	started time.Time
	span    trace.Span
}

func (m *Manager) InstallDatabase(db *gorm.DB) error {
	if !m.cfg.Enabled || (!m.cfg.DatabaseTracesEnabled && !m.cfg.MetricsEnabled) || db == nil {
		return nil
	}
	system := db.Dialector.Name()
	if system != "mysql" && system != "postgres" {
		system = "other"
	}
	for _, callbacks := range []struct {
		operation string
		before    func(string, func(*gorm.DB)) error
		after     func(string, func(*gorm.DB)) error
	}{
		{"SELECT", db.Callback().Query().Before("*").Register, db.Callback().Query().After("*").Register},
		{"INSERT", db.Callback().Create().Before("*").Register, db.Callback().Create().After("*").Register},
		{"UPDATE", db.Callback().Update().Before("*").Register, db.Callback().Update().After("*").Register},
		{"DELETE", db.Callback().Delete().Before("*").Register, db.Callback().Delete().After("*").Register},
		{"ROW", db.Callback().Row().Before("*").Register, db.Callback().Row().After("*").Register},
		{"EXEC", db.Callback().Raw().Before("*").Register, db.Callback().Raw().After("*").Register},
	} {
		operation := callbacks.operation
		attrs := []attribute.KeyValue{attribute.String("db.system.name", system), attribute.String("db.operation.name", operation)}
		if err := callbacks.before("postfixadmin:telemetry:before", func(tx *gorm.DB) {
			ctx := tx.Statement.Context
			if tx.DryRun || ctx.Value(databaseRequestKey{}) != true {
				return
			}
			state := &databaseState{parent: ctx, started: time.Now()}
			if m.cfg.DatabaseTracesEnabled && m.traces != nil && trace.SpanContextFromContext(ctx).IsValid() {
				ctx, state.span = m.traces.Tracer("go-postfixadmin/database").Start(ctx, "DB "+operation,
					trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(attrs...))
			}
			tx.Statement.Context = context.WithValue(ctx, databaseStateKey{}, state)
		}); err != nil {
			return fmt.Errorf("register database telemetry callback")
		}
		if err := callbacks.after("postfixadmin:telemetry:after", func(tx *gorm.DB) {
			state, ok := tx.Statement.Context.Value(databaseStateKey{}).(*databaseState)
			if !ok || tx.DryRun {
				return
			}
			errorType := "none"
			if tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				errorType = "database_error"
				if state.span != nil {
					state.span.SetStatus(codes.Error, errorType)
					state.span.SetAttributes(attribute.String("error.type", errorType))
				}
			}
			if m.metrics != nil {
				options := metric.WithAttributes(append(append([]attribute.KeyValue{}, attrs...), attribute.String("error.type", errorType))...)
				m.dbOperations.Add(tx.Statement.Context, 1, options)
				m.dbDuration.Record(tx.Statement.Context, time.Since(state.started).Seconds(), options)
			}
			if state.span != nil {
				state.span.End()
			}
			tx.Statement.Context = state.parent
		}); err != nil {
			return fmt.Errorf("register database telemetry callback")
		}
	}
	if m.metrics != nil {
		return m.installPoolMetrics(db)
	}
	return nil
}

func (m *Manager) installPoolMetrics(db *gorm.DB) error {
	pool, err := db.DB()
	if err != nil {
		return fmt.Errorf("initialize database pool metrics")
	}
	meter := m.metrics.Meter("go-postfixadmin/database")
	connections, err := meter.Int64ObservableGauge("db.client.connection.count", metric.WithUnit("{connection}"))
	if err != nil {
		return err
	}
	_, err = meter.RegisterCallback(func(_ context.Context, observer metric.Observer) error {
		stats := pool.Stats()
		observer.ObserveInt64(connections, int64(stats.InUse), metric.WithAttributes(attribute.String("db.client.connection.state", "used")))
		observer.ObserveInt64(connections, int64(stats.Idle), metric.WithAttributes(attribute.String("db.client.connection.state", "idle")))
		return nil
	}, connections)
	return err
}
