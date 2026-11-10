package server

import (
	"net/http/httptest"
	"testing"
)

func TestClientIPExtractor(t *testing.T) {
	cases := []struct {
		name, remote, xff string
		proxies           []string
		want              string
	}{
		{"direct client", "203.0.113.9:4000", "", nil, "203.0.113.9"},
		{"public peer cannot spoof", "203.0.113.9:4000", "198.51.100.1", nil, "203.0.113.9"},
		{"local reverse proxy", "127.0.0.1:4000", "198.51.100.1", nil, "198.51.100.1"},
		{"LAN client cannot spoof", "192.168.1.50:4000", "198.51.100.1", nil, "192.168.1.50"},
		{"configured container network proxy", "172.18.0.2:4000", "198.51.100.1", []string{"172.18.0.0/16"}, "198.51.100.1"},
		{"configured public proxy", "203.0.113.9:4000", "198.51.100.1", []string{"203.0.113.0/24"}, "198.51.100.1"},
		{"configured single address", "203.0.113.9:4000", "198.51.100.1", []string{"203.0.113.9"}, "198.51.100.1"},
	}
	for _, tc := range cases {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = tc.remote
		if tc.xff != "" {
			r.Header.Set("X-Forwarded-For", tc.xff)
		}
		if got := clientIPExtractor(tc.proxies)(r); got != tc.want {
			t.Errorf("%s: got %s, want %s", tc.name, got, tc.want)
		}
	}
}
