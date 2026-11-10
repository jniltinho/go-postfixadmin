package utils

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

const sample = `# Go-PostfixAdmin
[server]
port = 8080
swagger_enable = true # Swagger UI on /swagger/

[smtp]
server  = "localhost"   # SMTP relay
port    = 25
type    = "plain" # type: plain | tls | starttls

[vacation]
enabled = true
`

func parse(t *testing.T, doc string) *viper.Viper {
	t.Helper()
	v := viper.New()
	v.SetConfigType("toml")
	if err := v.ReadConfig(strings.NewReader(doc)); err != nil {
		t.Fatalf("result is not valid TOML: %v\n%s", err, doc)
	}
	return v
}

func TestSetTOMLValue(t *testing.T) {
	out, err := SetTOMLValue(sample, "smtp", "server", `"mail.example.com"`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `server  = "mail.example.com" # SMTP relay`) {
		t.Fatalf("spacing or comment lost:\n%s", out)
	}
	out, _ = SetTOMLValue(out, "server", "swagger_enable", "false")
	out, _ = SetTOMLValue(out, "smtp", "type", `"starttls"`)
	out, _ = SetTOMLValue(out, "vacation", "missing_key", "true")
	out, _ = SetTOMLValue(out, "ui", "default_theme", `"xp"`)
	v := parse(t, out)
	checks := map[string]any{
		"smtp.server": "mail.example.com", "server.swagger_enable": false, "smtp.type": "starttls",
		"vacation.missing_key": true, "ui.default_theme": "xp", "server.port": int64(8080), "smtp.port": int64(25),
	}
	for k, want := range checks {
		if got := v.Get(k); got != want {
			t.Errorf("%s = %#v, want %#v", k, got, want)
		}
	}
	for _, keep := range []string{"# Go-PostfixAdmin", "# Swagger UI on /swagger/", "# type: plain | tls | starttls"} {
		if !strings.Contains(out, keep) {
			t.Errorf("comment %q lost", keep)
		}
	}
	if strings.Index(out, "missing_key") > strings.Index(out, "[ui]") {
		t.Error("a missing key must be added inside its own section")
	}
}

func TestSetTOMLValueQuotesAndHash(t *testing.T) {
	doc := "[smtp]\nserver = \"a#b\" # c\n"
	out, err := SetTOMLValue(doc, "smtp", "server", `"x \"y\" #z"`)
	if err != nil {
		t.Fatal(err)
	}
	if got := parse(t, out).GetString("smtp.server"); got != `x "y" #z` {
		t.Fatalf("got %q\n%s", got, out)
	}
	if _, err := SetTOMLValue(doc, "smtp", "server", "\"a\"\nevil = 1"); err == nil {
		t.Fatal("multi-line values must be rejected")
	}
}
