package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go-postfixadmin/internal/auth"

	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
)

func TestAuthRefreshRejectsIdleSession(t *testing.T) {
	viper.Set("server.jwt_access_ttl", "1ms")
	viper.Set("server.session_idle_timeout", "1ms")
	t.Cleanup(func() {
		viper.Set("server.jwt_access_ttl", nil)
		viper.Set("server.session_idle_timeout", nil)
	})
	tok, err := auth.GenerateRefreshToken(auth.TokenParams{Username: "admin@example.com", Type: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1100 * time.Millisecond) // iat has one-second resolution

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: tok})
	rec := httptest.NewRecorder()
	if err := (&Handler{}).AuthRefresh(echo.New().NewContext(req, rec)); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusUnauthorized || !strings.Contains(rec.Body.String(), "inactivity") {
		t.Fatalf("idle refresh: %d %s", rec.Code, rec.Body)
	}
	if c := rec.Header().Get("Set-Cookie"); !strings.Contains(c, "refresh_token=;") {
		t.Fatalf("refresh cookie not cleared: %q", c)
	}
}

func TestRefreshTokenKeepsSessionExpiry(t *testing.T) {
	exp := time.Now().Add(2 * time.Hour).Truncate(time.Second)
	tok, err := auth.GenerateRefreshToken(auth.TokenParams{Username: "u", Type: "admin", Expires: exp})
	if err != nil {
		t.Fatal(err)
	}
	cl, err := auth.ValidateToken(tok)
	if err != nil || !cl.ExpiresAt.Time.Equal(exp) {
		t.Fatalf("expiry %v, want %v (err %v)", cl.ExpiresAt, exp, err)
	}
}
