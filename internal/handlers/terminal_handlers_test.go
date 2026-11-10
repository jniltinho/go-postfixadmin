package handlers

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"go-postfixadmin/internal/auth"
	"go-postfixadmin/internal/rbac"

	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

func TestTerminalAllowed(t *testing.T) {
	viper.Set("rbac.enabled", false)
	t.Cleanup(func() { viper.Set("rbac.enabled", false) })
	cases := []struct {
		name string
		cl   *auth.Claims
		rbac bool
		want bool
	}{
		{"nil", nil, false, false},
		{"mailbox token", &auth.Claims{Type: "mailbox", Superadmin: true}, false, false},
		{"domain admin", &auth.Claims{Type: "admin"}, false, false},
		{"superadmin", &auth.Claims{Type: "admin", Superadmin: true}, false, true},
		{"superadmin without settings:write under rbac", &auth.Claims{Type: "admin", Superadmin: true}, true, false},
		{"superadmin with settings:write under rbac", &auth.Claims{Type: "admin", Superadmin: true, Permissions: []string{rbac.PermSettingsWrite}}, true, true},
	}
	for _, tc := range cases {
		viper.Set("rbac.enabled", tc.rbac)
		if got := terminalAllowed(tc.cl); got != tc.want {
			t.Errorf("%s: got %v", tc.name, got)
		}
	}
}

func TestSameOrigin(t *testing.T) {
	for origin, want := range map[string]bool{
		"":                        false,
		"null":                    false,
		"http://panel:8080":       true,
		"https://panel:8080":      true,
		"http://evil.example":     false,
		"http://panel:9090":       false,
		"javascript://panel:8080": false,
	} {
		r := httptest.NewRequest("GET", "http://panel:8080/api/v1/system/terminal/ws", nil)
		r.Host = "panel:8080"
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		if got := sameOrigin(r); got != want {
			t.Errorf("origin %q: got %v", origin, got)
		}
	}
}

func TestCheckHostKeyPinned(t *testing.T) {
	newKey := func() ssh.PublicKey {
		pub, _, _ := ed25519.GenerateKey(rand.Reader)
		k, _ := ssh.NewPublicKey(pub)
		return k
	}
	k1, k2 := newKey(), newKey()
	line := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(k1)))
	if err := checkHostKey(nil, "h:22", line, "", k1, nil); err != nil {
		t.Fatalf("pinned key must pass: %v", err)
	}
	var ce *codeError
	if err := checkHostKey(nil, "h:22", line, ssh.FingerprintSHA256(k2), k2, nil); !errors.As(err, &ce) || ce.code != codeHostChanged {
		t.Fatalf("other key must be refused as changed even when 'trusted': %v", err)
	}
	if err := checkHostKey(nil, "h:22", "", "", k2, nil); !errors.As(err, &ce) || ce.code != codeHostUnknown || ce.fingerprint != ssh.FingerprintSHA256(k2) {
		t.Fatalf("unpinned key must ask for confirmation: %v", err)
	}
	if err := checkHostKey(nil, "h:22", "", ssh.FingerprintSHA256(k1), k2, nil); !errors.As(err, &ce) || ce.code != codeHostUnknown {
		t.Fatalf("a confirmation of another fingerprint must not pin: %v", err)
	}
}

func TestHostKeyAlgos(t *testing.T) {
	if got := hostKeyAlgos(""); got[0] != ssh.KeyAlgoED25519 {
		t.Fatalf("unpinned: %v", got)
	}
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	k, _ := ssh.NewPublicKey(pub)
	if got := hostKeyAlgos(string(ssh.MarshalAuthorizedKey(k))); len(got) != 1 || got[0] != ssh.KeyAlgoED25519 {
		t.Fatalf("pinned ed25519: %v", got)
	}
}
