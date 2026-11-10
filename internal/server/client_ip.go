package server

import (
	"log/slog"
	"net"
	"strings"

	"github.com/labstack/echo/v5"
)

// clientIPExtractor takes the client address from X-Forwarded-For only when the connection comes from loopback (a
// reverse proxy on the same host) or from a range in server.trusted_proxies (e.g. a proxy in a Docker network); any
// other peer, LAN clients included, is identified by its socket address, so it cannot spoof its IP to the login rate
// limiter or the audit log. Echo >= 5.4 no longer reads the header by default.
func clientIPExtractor(proxies []string) echo.IPExtractor {
	opts := []echo.TrustOption{echo.TrustLinkLocal(false), echo.TrustPrivateNet(false)}
	for _, p := range proxies {
		cidr := strings.TrimSpace(p)
		if !strings.Contains(cidr, "/") {
			if strings.Contains(cidr, ":") {
				cidr += "/128"
			} else {
				cidr += "/32"
			}
		}
		_, n, err := net.ParseCIDR(cidr)
		if err != nil {
			slog.Warn("ignoring invalid server.trusted_proxies entry", "value", p)
			continue
		}
		opts = append(opts, echo.TrustIPRange(n))
	}
	return echo.ExtractIPFromXFFHeader(opts...)
}
