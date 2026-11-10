1. **[P1] LAN clients can still spoof their IP** — `internal/server/client_ip.go:16–33`: Echo’s defaults trust every private/link-local peer, even with `trusted_proxies=[]`. A direct client at `192.168.1.50` can rotate `X-Forwarded-For` to bypass login throttling and falsify audit IPs. **Fix:** disable blanket trust and explicitly allow configured proxy addresses. [Echo trust options](https://echo.labstack.com/guide/ip-address/)

**Other fix areas**
- Docker `TARGETARCH`: none.
- Host-key pin race: none.
- `/system/services` mailbox-token exclusion: none.
- Terminal reconnect after unmount: none.
- Vite WebSocket proxy: none.
- Theme initialization race: none.
**New i18n**
- `lang.ts` and Settings dictionary: none.
- `SettingsPage.vue` and `TerminalTab.vue`: none.
- `HeaderBar.vue` and `SidebarNav.vue`: none.
- Both login pages and `UserDashboardPage.vue`: none.
Vue typecheck and in-memory locale/theme/unmount checks passed. No files edited.  
Not checked: Go tests (blocked by read-only `/tmp`), Docker builds, or live SSH/proxy/browser sessions.
143.735
