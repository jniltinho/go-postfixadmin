## Context

- Frontend: Vue 3 + TS + Tailwind v4 + Vite 8, built to `web/dist` and embedded with `//go:embed`. Colors are hard-coded (520 hex values in 33 files, plus Tailwind palette classes such as `bg-white` x100, `text-gray-500` x52). `MainLayout.vue` uses a fixed `240px / 56px` grid and `#ebf2fe` background.
- ad-manager (same author) already solved theming: a role contract of ~318 CSS variables in `style.css`, one file per theme family under `src/themes/`, `<html data-theme="id">`, and a Go registry (`internal/server/themes.go`) with `id`, `label`, `scheme`. Its XP Luna theme is `classic-blue` in `classics.css`.
- Backend: Echo v5.0.4, `echo-contrib` v0.50.0 (retracted), Go 1.26. Local toolchain already has Go 1.27.3.
- Release CI builds the binary, `.deb` and `.rpm`; no image is pushed. Baseline screenshots are in `evidence/`.

## Goals / Non-Goals

**Goals**
- Theming with zero visual regression for the current look (`default`).
- All ad-manager themes plus a theme named XP available in admin UI and mailbox portal.
- Usable, proportional layout at 1920x1080, 2560x1440 and 3840x2160 (DPR 1), still working at 1366x768.
- Toolchain on current stable: Go 1.27, Echo v5.4.x, Vite+.
- One-command install from GHCR.

**Non-Goals**
- Per-user theme stored in the database (browser `localStorage` + server default is enough; revisit if users ask).
- Controlling services (start/stop/restart) from Settings, or several terminal targets.
- Theme editor or custom uploaded themes.
- Redesigning page flows, adding pages or changing the REST contract beyond one additive field.
- Removing legacy `web/templates` (separate change).
- Windows/macOS container images.

## Decisions

### D1. Copy ad-manager's role contract, do not invent a new one
Reuse the variable names (`--color-page`, `--color-surface`, `--color-fg`, `--color-action`, `--color-line`, `--shadow-*`, `--bevel-*` ...) and the theme files as-is, so future themes can be copied between projects without edits. Only roles actually used by go-postfixadmin components are kept in `style.css`; unused roles are dropped from the copied theme files during the port to keep the CSS small.
*Alternative*: Tailwind `dark:` variants only — rejected, it gives two looks, not a catalog.

### D2. `default` = today's neo-brutalist look
The current colors become the `default` theme values, so the migration is a pure refactor before new themes are added. Verified by pixel diff of `default` against the baseline screenshots (tolerance for antialiasing only).

### D3. XP theme
Port ad-manager's `classic-blue` as id `xp`, label "Windows XP". Keep `classic-gray` as "Windows 98". Tahoma font family only inside these two themes; sizes stay on the responsive scale of D5 (no fixed 11px), so the high-resolution minimums hold for XP too. Contrast fixes from ad-manager (AA) are kept.

### D4. Theme selection and persistence
- Registry: a typed array in `frontend/src/themes/index.ts` (`{ id, label, scheme }`) — single source for the selectors; ids must match a `[data-theme='id']` block (checked by a unit test). Selectors are `[data-theme]`, not `:root[data-theme]`, so a preview card in Settings can carry another theme; `themes/default.css` repeats the role defaults for that (a test keeps it equal to `style.css`).
- Where: Settings › Appearance (cards with live previews) in the admin UI; a compact select in the mailbox portal header.
- Server default: `[ui] default_theme = "default"` (Viper key `ui.default_theme`, env `UI_DEFAULT_THEME`), returned by `GET /api/v1/version` as `default_theme`. Unknown ids fall back to `default` with a `slog.Warn` at startup.
- Precedence: `localStorage['theme']` > server default > `default`.
- `index.html` inline script sets `data-theme` and `color-scheme` from `localStorage` (`theme`, `theme-scheme`, both written by the store) before the bundle loads, so neither colors nor native controls flash.
- If `/api/v1/version` fails, the SPA keeps the stored choice or `default`. An unknown `ui.default_theme` is not validated on the server (the id list lives in the frontend); clients fall back to `default`.
- Pinia store `theme.ts` sets `document.documentElement.dataset.theme`, `style.colorScheme` and `<meta name="color-scheme">`.

### D5. High-resolution layout
- Root font size `clamp(16px, 0.5vw + 6.4px, 24px)` → 16 px up to 1920 (FullHD renders exactly as before), 19.2 px at 2560, 24 px at 3840; Tailwind `rem` utilities scale with it.
- Every scoped sizing value (`font-size`, padding, margin, gap, width/height, letter-spacing…) is converted from `px` to `rem` (px/16), including Tailwind arbitrary values (`text-[10px]` → `text-[0.625rem]`) and the `Icon` component's size; borders, shadow offsets and transforms stay in `px`. The sidebar (15rem) and header (3.5rem) scale through the same conversion, so no separate layout tokens or breakpoints are needed.
- Login column `max-width: max(28rem, 23vw)`; portal content `max-w-[max(72rem,60vw)]` (both identical to before at FullHD).
- No `zoom` / `transform: scale` hacks.

### D6. Vite+
Use `vite-plus` with the same `overrides` pattern as ad-manager (`vite` → `@voidzero-dev/vite-plus-core`), removing the direct `vite` devDependency first (npm rejects an override that conflicts with a direct dependency, `EOVERRIDE`). Scripts: `dev: vp dev`, `build: vue-tsc -b && vp build`, `check: vp check` (oxfmt with `singleQuote`, no semicolons, width 120 + oxlint), `test: vp test run`. `vite.config.ts` imports `defineConfig` from `vite-plus`; `outDir`, proxy and `API_BASE` define stay. The tracked generated siblings `vite.config.js` / `vite.config.d.ts` are deleted and git-ignored (Vite resolves `.js` before `.ts`, so a stale `.js` would shadow the new config); `tsconfig.node.json` emits into `node_modules/.tmp` (`outDir`), since a composite project referenced by `tsc -b` cannot use `noEmit`; tests are excluded from `tsconfig.app.json`. The Makefile `frontend` target keeps calling `npm run build`, so CI and Dockerfile need no `vp` global install; the Docker Node stage moves to `node:24-alpine` (Vite+ needs Node ≥ 22.18).

### D7. Go / Echo upgrade
`go 1.27.0` in `go.mod` (toolchain directive not added). `go get github.com/labstack/echo/v5@latest github.com/labstack/echo-contrib/v5@latest` and rewrite the `echo-contrib/session` imports to `echo-contrib/v5/session` (v0.50.x targets Echo v4; v0.50.0 was retracted for that reason), then `go get -u=patch ./...`. Run `go vet` and `go test ./...`; fix compile errors from v5.x API changes (the handlers use `*echo.Context`). Follow AGENTS.md: re-run `go mod download` after `go mod tidy`.

### D8. GHCR image
New workflow `.github/workflows/docker.yml` on `push: tags: ['v*']` (+ `workflow_dispatch`): `docker/setup-qemu-action`, `docker/setup-buildx-action`, `docker/login-action` with `GITHUB_TOKEN`, `docker/metadata-action` (tags: semver `{{version}}` and `{{major}}.{{minor}}` on tags, `latest` only for stable semver tags without pre-release, `sha-<short>` only for `workflow_dispatch`; OCI labels incl. `org.opencontainers.image.source` and `org.opencontainers.image.description`, plus `annotations` with `index:org.opencontainers.image.description` so the multi-arch manifest index carries the description GHCR shows); the workflow triggers on `v*.*.*` semver tags only, `docker/build-push-action` with `platforms: linux/amd64,linux/arm64`, `target` = final stage, build args `VERSION`/`GIT_COMMIT`, GHA cache. Permissions: `contents: read`, `packages: write`. The Node and Go stages run on `$BUILDPLATFORM` and Go cross-compiles with `GOOS/GOARCH=$TARGETOS/$TARGETARCH` (no QEMU for the slow part); UPX packs the arm64 binary on the build host (verified locally with UPX 5). After the first push the package visibility is set to public once (GitHub UI, owner action) and anonymous `docker pull` is verified.

### D9. Validation
- `agent-browser`: script under `tests/visual/` that logs in against the docker-compose stack and captures login, dashboard, mailboxes, a modal and the user portal for every theme x {1920x1080, 2560x1440, 3840x2160}; checks no horizontal scroll and login card width ≥ 22% of the viewport at QHD/4K.
- Contrast: reuse ad-manager's AA rule — automated check of `--color-fg` on `--color-page`/`--color-surface` ≥ 4.5:1 per theme.
- Codex: `codex exec` review of the OpenSpec change before implementation and of the final diff before merge; findings tracked in `tasks.md`.

### D9b. Client IP after the Echo 5.4 upgrade
Echo 5.0.4's `RealIP()` read `X-Forwarded-For` from anyone (spoofable); 5.4.4 uses the socket address, which behind a proxy would put every client under one login-rate-limit bucket. The server now sets `e.IPExtractor = echo.ExtractIPFromXFFHeader(...)` trusting only loopback plus `server.trusted_proxies` (IPs or CIDRs); LAN and other peers cannot spoof their address.

### D10. Settings › Services
`GET /api/v1/system/services` (permission `settings:read`) runs `systemctl show <unit> --property=LoadState,ActiveState,SubState,UnitFileState,ActiveEnterTimestamp` for a fixed allowlist (`postfix.service`, `dovecot.service`) with a 5 s timeout — the same source pf-report's Manager page uses, never the colored `systemctl status` text. A unit with `LoadState=not-found` is reported as not installed; no `systemctl` (e.g. in a container) is reported as unavailable. Read-only: start/stop/restart would need root and is out of scope.

### D11. Settings › Terminal (web SSH)
Ported from ad-manager `internal/server/ssh.go` / `pages/Terminal.vue`:
- Off unless `[terminal] enabled = true`; target `terminal.host`:`terminal.port` (default `127.0.0.1:22`), `terminal.idle_timeout` (30m).
- `GET /api/v1/system/terminal/ws` is outside the JWT group because browsers cannot send headers on a WebSocket: the handler first checks that `Origin` matches `Host`, then the first message carries the access token, validated as an `admin` token of a superadmin (with RBAC on, also `settings:write`) before anything is dialed. Info and host-key reset are normal JWT routes with `RequirePermission(settings:write)` plus the same superadmin check.
- SSH via `golang.org/x/crypto/ssh`: password, keyboard-interactive or private key (+ passphrase); PTY `xterm-256color`; credential and token dropped from memory after the handshake; 4 terminals per admin; 64 KiB frames; 30 s pings; idle timeout counted from the last keystroke.
- Host key: table `terminal_host_key` (host → authorized_keys line). Unknown key → the browser shows `type SHA256:…` and only a reconnect carrying that exact fingerprint pins it (insert-if-absent); a different key than the pinned one is refused (`hostkey_changed`) until a superadmin resets it.
- Audit (`log` table): `terminal_hostkey_trust`, `terminal_open`, `terminal_open_failed <code>`, `terminal_close <duration>`, `terminal_hostkey_reset`. Commands typed are not recorded (the server's own login records are).
- Browser: xterm.js 6 + fit addon; colors from `--term-*` roles (`themes/terminal.css`: classic xterm palette by default, page colors for dark themes, cmd.exe for XP/98, phosphor for Matrix) re-applied live when the theme changes; the tab stays mounted (`v-show`) so a session survives switching tabs.

### D12. Settings at AAA
Secondary text uses `--color-fg-2` instead of `--color-fg-muted`; status is told by icon + colored edge with `--color-fg` text; table heads and the primary button are inverted (`--color-surface` on `--color-fg`) instead of white on the action blue (3.7:1); visible 3px focus outline. Verified by an in-page checker over every text node of the Settings tabs (≥ 7:1, ≥ 4.5:1 for large text) in default, xp, classic-gray, breeze-dark, matrix and high-contrast.

## Risks / Trade-offs

- **Web terminal = shell access from the browser** → off by default, superadmin-only, same-origin, pinned host key, audited, idle timeout, credentials not stored; documented as opt-in in the setup guide.

- **Large mechanical refactor (33 files)** → do it in batches per page folder, screenshot-diff `default` after each batch.
- **Vite+ is young (1.x)** → keep `vite` fallback trivial: revert scripts + config import; pin exact `vite-plus` version.
- **Echo minor releases may change middleware signatures** → upgrade in its own commit, before UI work, with full test run.
- **Multi-arch + UPX** may fail under QEMU → guard by `TARGETARCH`.
- **Go 1.27 as minimum** drops builds with 1.26 → documented in README/DEVELOPMENT.
- **localStorage-only preference** does not follow the user across browsers → accepted (non-goal).

## Migration Plan

0. Capture baselines (login, dashboard, mailboxes, a modal, user portal) at FHD/QHD/4K with seeded data; add HTTP regression tests for login, refresh, an RBAC-guarded route and a public route.
1. Toolchain commit (Go/Echo/deps), tests green.
2. Vite+ commit, `make build` green, bundle served.
3. Role contract + `default` theme, pixel diff = baseline (sizing unchanged in this step).
4. Theme catalog + XP + selector + config key + i18n.
5. High-resolution layout.
6. GHCR workflow; test with `workflow_dispatch` (pushes `sha-<short>` only, never `latest`); set package visibility to public.
7. Docs + screenshots; Codex review; release.

Rollback: each step is a separate commit; revert the step. Image tags are immutable per version.

## Open Questions

- Should the theme also be stored per admin in the DB later? (default answer: no, out of scope)
- Keep `classic-gray` (Windows 98) alongside XP, or ship only XP from the classics family? (proposal keeps both)
