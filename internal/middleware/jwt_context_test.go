package middleware

import (
	"context"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestAPIKeyLookupUsesRequestContext(t *testing.T) {
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
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var observed context.Context
	if err := db.Callback().Query().Before("gorm:query").Register("test:request_context", func(tx *gorm.DB) {
		observed = tx.Statement.Context
	}); err != nil {
		t.Fatal(err)
	}
	if got := resolveAPIKeyClaims(ctx, db, "fake-api-key"); got != nil {
		t.Fatal("canceled lookup must not authenticate")
	}
	if observed != ctx {
		t.Fatal("API key lookup lost request context")
	}
	if db.Statement.Context.Err() != nil {
		t.Fatal("API key lookup mutated shared database session")
	}
	if got := resolveAPIKeyClaims(ctx, nil, "fake-api-key"); got != nil {
		t.Fatal("nil database must not authenticate")
	}
}
