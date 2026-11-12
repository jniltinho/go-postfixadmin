package middleware

import "testing"

func TestHostMatches(t *testing.T) {
	for _, tc := range []struct {
		header string
		want   bool
	}{
		{"https://panel.example.com", true},
		{"https://panel.example.com/login?next=/", true},
		{"https://PANEL.example.com", true},
		{"https://panel.example.com.attacker.tld", false},
		{"https://panel.example.com:8443", false},
		{"https://evil.tld/https://panel.example.com", false},
	} {
		if got := hostMatches(tc.header, normalizeHost("https://panel.example.com/")); got != tc.want {
			t.Errorf("hostMatches(%q) = %v, want %v", tc.header, got, tc.want)
		}
	}
}
