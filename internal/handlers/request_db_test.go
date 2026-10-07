package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-postfixadmin/internal/models"

	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestRequestDBCancellationAndIsolation(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: "test:test@tcp(127.0.0.1:1)/test", SkipInitializeWithVersion: true,
	}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	h := &Handler{DB: db}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	e := echo.New()
	c := e.NewContext(httptest.NewRequest("GET", "/", nil).WithContext(ctx), httptest.NewRecorder())
	requestDB := h.requestDB(c)
	if requestDB.Statement.Context != ctx {
		t.Fatal("request database lost the request context")
	}
	var admin models.Admin
	if err := requestDB.First(&admin).Error; !errors.Is(err, context.Canceled) {
		t.Fatalf("query error = %v, want cancellation", err)
	}
	if h.DB != db || db.Statement.Context.Err() != nil {
		t.Fatal("request mutated the shared database session")
	}
	if got := (&Handler{}).requestDB(c); got != nil {
		t.Fatal("nil database must remain nil")
	}
}

func TestRequestDBMutationCompletesAfterCancellation(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			db, err := gorm.Open(mysql.New(mysql.Config{
				DSN: "test:test@tcp(127.0.0.1:1)/test", SkipInitializeWithVersion: true,
			}), &gorm.Config{DisableAutomaticPing: true, DryRun: true})
			if err != nil {
				t.Fatal(err)
			}
			sqlDB, err := db.DB()
			if err != nil {
				t.Fatal(err)
			}
			defer sqlDB.Close()
			spanContext := trace.NewSpanContext(trace.SpanContextConfig{
				TraceID: trace.TraceID{1}, SpanID: trace.SpanID{2}, TraceFlags: trace.FlagsSampled,
			})
			ctx, cancel := context.WithCancel(trace.ContextWithSpanContext(context.Background(), spanContext))
			defer cancel()
			operations := 0
			if err := db.Callback().Raw().Before("gorm:raw").Register("test:complete_mutation", func(tx *gorm.DB) {
				if err := tx.Statement.Context.Err(); err != nil {
					tx.AddError(err)
					return
				}
				if !trace.SpanContextFromContext(tx.Statement.Context).Equal(spanContext) {
					t.Error("mutation lost tracing context")
				}
				operations++
				if operations == 1 {
					cancel()
				}
			}); err != nil {
				t.Fatal(err)
			}
			h := &Handler{DB: db}
			c := echo.New().NewContext(httptest.NewRequest(method, "/", nil).WithContext(ctx), httptest.NewRecorder())
			for range 2 {
				if err := h.requestDB(c).Exec("UPDATE admins SET active = ? WHERE username = ?", true, "fake-admin").Error; err != nil {
					t.Fatalf("mutation interrupted: %v", err)
				}
			}
			if ctx.Err() != context.Canceled || operations != 2 {
				t.Fatalf("request cancellation = %v, completed operations = %d", ctx.Err(), operations)
			}
		})
	}
}
