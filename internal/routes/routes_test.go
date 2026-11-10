package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-postfixadmin/internal/auth"
	"go-postfixadmin/internal/handlers"
	"go-postfixadmin/internal/rbac"

	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newTestServer(t *testing.T) *echo.Echo {
	t.Helper()
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
	t.Cleanup(func() { sqlDB.Close() })
	e := echo.New()
	RegisterRoutes(e, &handlers.Handler{DB: db}, nil)
	return e
}

func adminToken(t *testing.T, perms ...string) string {
	t.Helper()
	tok, err := auth.GenerateAccessToken(auth.TokenParams{Username: "a@example.com", Type: "admin", Permissions: perms})
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func mailboxToken(t *testing.T) string {
	t.Helper()
	tok, err := auth.GenerateAccessToken(auth.TokenParams{Username: "ana@example.com", Type: "mailbox"})
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func TestRouteStatusRegression(t *testing.T) {
	viper.Set("rbac.enabled", true)
	t.Cleanup(func() { viper.Set("rbac.enabled", false) })
	e := newTestServer(t)

	cases := []struct {
		name, method, path, body, token string
		want                            int
		notIn                           []int
	}{
		{name: "public version", method: http.MethodGet, path: "/api/v1/version", want: http.StatusOK},
		{name: "me without token", method: http.MethodGet, path: "/api/v1/auth/me", want: http.StatusUnauthorized},
		{name: "invalid token", method: http.MethodGet, path: "/api/v1/domains", token: "garbage", want: http.StatusUnauthorized},
		{name: "login missing fields", method: http.MethodPost, path: "/api/v1/auth/login", body: `{}`, want: http.StatusBadRequest},
		{name: "refresh without cookie", method: http.MethodPost, path: "/api/v1/auth/refresh", body: `{}`, want: http.StatusUnauthorized},
		{name: "rbac without permission", method: http.MethodGet, path: "/api/v1/domains", token: adminToken(t), want: http.StatusForbidden},
		{name: "services without settings:read", method: http.MethodGet, path: "/api/v1/system/services", token: adminToken(t, rbac.PermDomainsRead), want: http.StatusForbidden},
		{name: "services with a mailbox token", method: http.MethodGet, path: "/api/v1/system/services", token: mailboxToken(t), want: http.StatusForbidden},
		{name: "services with settings:read", method: http.MethodGet, path: "/api/v1/system/services", token: adminToken(t, rbac.PermSettingsRead), want: http.StatusOK},
		{name: "terminal ws disabled", method: http.MethodGet, path: "/api/v1/system/terminal/ws", want: http.StatusNotFound},
		{name: "terminal info for a domain admin", method: http.MethodGet, path: "/api/v1/system/terminal", token: adminToken(t, rbac.PermSettingsWrite), want: http.StatusForbidden},
		{name: "rbac with permission", method: http.MethodGet, path: "/api/v1/domains", token: adminToken(t, rbac.PermDomainsRead),
			notIn: []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if tc.want != 0 && rec.Code != tc.want {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tc.want, rec.Body)
			}
			for _, bad := range tc.notIn {
				if rec.Code == bad {
					t.Fatalf("status = %d must not be %d (body %s)", rec.Code, bad, rec.Body)
				}
			}
		})
	}
}

func TestTerminalWSRefusesCrossOrigin(t *testing.T) {
	viper.Set("terminal.enabled", true)
	t.Cleanup(func() { viper.Set("terminal.enabled", false) })
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/terminal/ws", nil)
	req.Header.Set("Origin", "http://evil.example")
	rec := httptest.NewRecorder()
	newTestServer(t).ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status %d, want 403", rec.Code)
	}
}

func TestVersionReportsDefaultTheme(t *testing.T) {
	viper.Set("ui.default_theme", "xp")
	t.Cleanup(func() { viper.Set("ui.default_theme", nil) })
	rec := httptest.NewRecorder()
	newTestServer(t).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/version", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"default_theme":"xp"`) {
		t.Fatalf("status %d body %s", rec.Code, rec.Body)
	}
}
