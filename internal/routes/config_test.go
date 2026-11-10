package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-postfixadmin/internal/auth"

	"github.com/spf13/viper"
)

func superToken(t *testing.T) string {
	t.Helper()
	tok, err := auth.GenerateAccessToken(auth.TokenParams{Username: "root@example.com", Type: "admin", Superadmin: true})
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func TestConfigAPI(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	orig := "[database]\npass = \"s3cret-db\"\n\n[smtp]\nserver = \"localhost\" # relay\nport   = 25\n\n[server]\nport = 8080\n"
	if err := os.WriteFile(path, []byte(orig), 0o640); err != nil {
		t.Fatal(err)
	}
	viper.Reset()
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(viper.Reset)
	e := newTestServer(t)
	do := func(method, body, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "/api/v1/system/config", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec
	}

	get := do(http.MethodGet, "", superToken(t))
	if get.Code != http.StatusOK || strings.Contains(get.Body.String(), "s3cret-db") {
		t.Fatalf("GET %d, secret leaked or failed: %s", get.Code, get.Body)
	}
	if !strings.Contains(get.Body.String(), `"key":"database.pass","kind":"string","editable":false,"secret":true,"set":true`) {
		t.Fatalf("secret should be reported as set without its value: %s", get.Body)
	}

	// TLS in effect: cert + key without ssl_enable switch it on (cmd/server.go); JWT lifetimes show their defaults
	viper.Set("server.ssl_cert", "/etc/ssl/panel.crt")
	viper.Set("server.ssl_key", "/etc/ssl/panel.key")
	eff := do(http.MethodGet, "", superToken(t)).Body.String()
	for _, want := range []string{`"key":"server.ssl_enable","kind":"bool","value":true`, `"key":"server.ssl_cert","kind":"string","value":"/etc/ssl/panel.crt"`, `"key":"server.jwt_access_ttl","kind":"duration","value":"15m"`, `"key":"server.jwt_refresh_ttl","kind":"duration","value":"168h"`} {
		if !strings.Contains(eff, want) {
			t.Fatalf("missing %s in %s", want, eff)
		}
	}

	if rec := do(http.MethodGet, "", adminToken(t)); rec.Code != http.StatusForbidden {
		t.Fatalf("domain admin GET = %d, want 403", rec.Code)
	}
	if rec := do(http.MethodPut, `{"changes":{"server.port":9090}}`, superToken(t)); rec.Code != http.StatusBadRequest {
		t.Fatalf("read-only key = %d, want 400", rec.Code)
	}
	if rec := do(http.MethodPut, `{"changes":{"smtp.port":70000}}`, superToken(t)); rec.Code != http.StatusBadRequest {
		t.Fatalf("out of range = %d, want 400", rec.Code)
	}
	if rec := do(http.MethodPut, `{"changes":{"smtp.server":"evil\"\n[x]"}}`, superToken(t)); rec.Code != http.StatusBadRequest {
		t.Fatalf("injection = %d, want 400", rec.Code)
	}

	rec := do(http.MethodPut, `{"changes":{"smtp.server":"relay.example.com","smtp.port":587,"server.swagger_enable":false}}`, superToken(t))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"restart_required":true`) {
		t.Fatalf("PUT %d: %s", rec.Code, rec.Body)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), `server = "relay.example.com" # relay`) || !strings.Contains(string(data), "port   = 587") {
		t.Fatalf("file not updated in place:\n%s", data)
	}
	if bak, _ := os.ReadFile(path + ".bak"); string(bak) != orig {
		t.Fatal("previous config not kept as .bak")
	}
	if st, _ := os.Stat(path); st.Mode().Perm() != 0o640 {
		t.Fatalf("mode changed to %v", st.Mode().Perm())
	}
	if viper.GetString("smtp.server") != "relay.example.com" || viper.GetInt("smtp.port") != 587 {
		t.Fatal("new values not effective in memory")
	}
	if rec := do(http.MethodPut, `{"changes":{"terminal.enabled":true}}`, superToken(t)); rec.Code != http.StatusOK || !viper.GetBool("terminal.enabled") {
		t.Fatalf("terminal.enabled = %d %s", rec.Code, rec.Body)
	}
	if rec := do(http.MethodPut, `{"changes":{"terminal.host":"10.0.0.9"}}`, superToken(t)); rec.Code != http.StatusBadRequest {
		t.Fatalf("terminal.host must stay read-only, got %d", rec.Code)
	}
}
