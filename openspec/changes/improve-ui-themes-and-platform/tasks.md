## 0. Pre-checks

- [x] 0.1 Confirm the local checkout is `git@github.com:jniltinho/go-postfixadmin.git` and `HEAD` equals `origin/main` (9b8cb88 at proposal time).
- [x] 0.2 Capture baseline login screenshots at 1920x1080, 2560x1440, 3840x2160 with `agent-browser` (`evidence/`).
- [x] 0.3 Query ai-memory for prior decisions (frontend build, single binary, Docker) and index the repo with codebase-memory-mcp.
- [x] 0.4 Run a Codex review of this change (`codex exec`, read-only) and address findings before implementation (16 findings over two passes, all incorporated; see `evidence/codex-review-*.md`).
- [x] 0.5 Capture baselines of login, dashboard, domains, mailboxes, a modal and the user portal at 1366/FHD/QHD/4K against a throwaway MariaDB with seeded data.
- [x] 0.6 Add HTTP regression tests (login, refresh, RBAC-guarded route with/without permission, public route) and get them green on Echo v5.0.4 (`internal/routes/routes_test.go`).

## 1. Toolchain: Go and Echo

- [x] 1.1 Set `go 1.27.0` in `go.mod`; `golang:1.27-alpine` in `Dockerfile` and `go-version: '1.27'` in `release.yml`.
- [x] 1.2 Upgrade `github.com/labstack/echo/v5` to v5.4.4 and replace `echo-contrib` with `echo-contrib/v5` (v5.0.1); `go get -u=patch ./...`; `go mod tidy` then `go mod download`.
- [x] 1.3 Fix API drift: only the `echo-contrib/session` imports (v0.50.x is for Echo v4) → `echo-contrib/v5/session`; no Echo v4 module left in the build graph.
- [x] 1.4 `go vet ./...`, `go test ./...` green; no retracted direct dependency; `make swagger` regenerated.

## 2. Toolchain: Vite+

- [x] 2.1 Remove the direct `vite` devDependency, add `vite-plus` 1.5.1 and the `vite` override; scripts `vp dev` / `vue-tsc -b && vp build` / `vp check` / `vp test run`.
- [x] 2.2 `vite.config.ts` uses `defineConfig` from `vite-plus` (+ `fmt` matching the code style); delete and git-ignore `vite.config.js`/`.d.ts`; `tsconfig.node.json` emits into `node_modules/.tmp`.
- [x] 2.3 Verify `npm run build` → `web/dist`, `npm run dev` proxy to the Go server, Docker frontend stage (`node:24-alpine`); `vp check --fix` applied once (formatting only).

## 3. Theme contract and default theme

- [x] 3.1 Role contract in `style.css` (`@theme static`) whose values are today's colors.
- [x] 3.2 Replace hard-coded colors and Tailwind palette classes with roles in 34 files (script mapping by CSS property); pixel diff of `default` vs baseline: 0 px on login/dashboard/domains/mailboxes, 37 px (modal) and 293 px (portal) from merging Tailwind red-500 into `#ef4444`.
- [x] 3.3 Unit test that fails on hex/rgba/palette classes outside `src/themes/` (`src/no-literal-colors.test.ts`).

## 4. Theme catalog, XP and selector

- [x] 4.1 Port the 19 ad-manager themes by resolving their role contract onto ours (`classic-blue` → `xp` "Windows XP" with Luna-blue title bars, `classic-gray` "Windows 98"); fill colors fitted to ≥ 4.5:1 against `on-action`.
- [x] 4.2 Registry `src/themes/index.ts` + tests: registry ↔ `[data-theme]` blocks, schemes, `default.css` ↔ `style.css`.
- [x] 4.3 `stores/theme.ts`, anti-flash script in `index.html`, Settings › Appearance (live previews) and the portal header select.
- [x] 4.4 `ui.default_theme` (Viper default, env `UI_DEFAULT_THEME`) in `GET /api/v1/version`, both TOML templates; `make swagger`.
- [x] 4.5 Portal select label in EN/PT/ES; admin SPA has no i18n layer, theme names stay proper names.
- [x] 4.6 Contrast: fg on page/surface and on-action on action/danger ≥ 4.5:1 for all 20 themes (76 pairs).
- [x] 4.7 Unit tests for the store precedence (stored > server default > default, unknown ids) and a Go test for `default_theme` on `/version`.

## 5. High-resolution layout

- [x] 5.1 Root font `clamp(16px, 0.5vw + 6.4px, 24px)` (unchanged at FullHD).
- [x] 5.2 Login column `max(28rem, 23vw)`, portal content `max(72rem, 60vw)`, Icon sizes in `rem`.
- [x] 5.3 `tests/visual/capture.sh` (agent-browser): every theme × login/dashboard/domains/mailboxes/settings/modal/portal × 1366/FHD/QHD/4K, with layout checks; logs in once (rate limiter stays on).
- [x] 5.4 Convert scoped `px` sizing (font, padding, margin, gap, sizes, Tailwind arbitrary values) to `rem`; borders/shadows stay `px`. FHD pixel diff ≈ 0 (only the new Settings menu entry).

## 6. GHCR image

- [x] 6.1 `.github/workflows/docker.yml`: `v*.*.*` tags + dispatch, QEMU, Buildx, ghcr.io login, metadata (semver, stable-only `latest`, `sha-` on dispatch, index description annotation), amd64+arm64, GHA cache, `packages: write`.
- [x] 6.2 Dockerfile: Node/Go stages on `$BUILDPLATFORM`, Go cross-compiles to `$TARGETARCH`, OCI labels; local amd64 build OK, arm64 cross-compile + UPX OK.
- [x] 6.3 `docker-compose.yml` uses `ghcr.io/jniltinho/go-postfixadmin:latest` (`build: .` kept as a comment).
- [ ] 6.4 Dry run via `workflow_dispatch` on GitHub (only `sha-<short>` tag); verify `docker pull` on amd64 and arm64 and `postfixadmin version`.
- [ ] 6.5 Owner sets the GHCR package visibility to public and links it to the repo; verify anonymous `docker pull`.

## 7. Documentation

- [x] 7.1 README: GHCR quick start and badge, themes & Settings section with screenshots, Go 1.27 / Vite+ / xterm.js stack.
- [x] 7.2 DEVELOPMENT (vp commands, theme contract, visual checks), FEATURES, CONTRIBUTING (theme rules, how to add a theme), AGENTS (Go 1.27, roles-only colors, rem sizes), setup guide (GHCR, `ui.default_theme`, Services, Terminal), `tests/visual/README.md`.
- [x] 7.3 New screenshots in `DOCUMENTS/screenshots/`: settings-themes, settings-services, settings-terminal(-hostkey), theme-xp, theme-matrix, theme-breeze-dark.

## 8. Settings › Services

- [x] 8.1 `GET /api/v1/system/services` (`settings:read`), fixed allowlist, `systemctl show` parsing with tests; not-installed and no-systemd states.
- [x] 8.2 Services tab with status badges (icon + text), refresh.

## 9. Settings › Terminal (web SSH)

- [x] 9.1 `terminal_host_key` model in `MigrateDB`; `[terminal]` config (enabled/host/port/idle_timeout) in both templates and Viper defaults.
- [x] 9.2 Handlers: info, host-key reset, WebSocket (same-origin, token in first message, superadmin + `settings:write`, 4 per admin, frame/ping/idle limits, `xterm-256color` PTY, pinned host key, audit log); unit tests for permissions, origin, pinning, algorithms; route tests (404 disabled, 403 cross-origin/non-superadmin).
- [x] 9.3 Terminal tab with xterm.js + fit, key file upload, fingerprint confirmation, forget key, classic xterm palette (black console, as in ad-manager) in every theme, kept mounted across tabs.
- [x] 9.4 End-to-end against a throwaway sshd container: fingerprint matches `ssh-keygen -lf`, `TERM=xterm-256color`, `tput colors` = 256, live theme switch, changed host key refused and logged.

## 10. Translations and language

- [x] 10a.1 `stores/lang.ts` (saved choice > browser > EN, Automatic, backend `lang` cookie) shared by login pages, admin UI and portal.
- [x] 10a.2 `pages/settings/i18n.ts` with EN/PT/ES for Settings, Terminal, dialogs, breadcrumb and menu entry; theme name words translated; tests for complete dictionaries and placeholders.
- [x] 10a.3 Remove the PT/EN/ES buttons from the admin header and both login pages; Language card in Settings › Appearance; portal select.

## 12. Follow-ups after the first deploy

- [x] 12.1 Menu: the selected item keeps its tab opening into the content; hover is a tint and left bar only (no box closed by the sidebar edge).
- [x] 12.2 Theme shape roles in `themes/shape.css`: subtle rounded corners (`--radius`, `--radius-sm`), soft shadows (`--shadow-1…7`, hard offsets kept for Default and High contrast) and lighter lines in the theme's hue (`--color-line-base` kept for form controls, 1px card frames) for every theme except Default; Windows 98 almost square.
- [x] 12.6 Services table with fixed columns and uniform badges; Configuration tab as two slim cards that fit FullHD/QHD without page or card scroll (scroll inside a card only on small screens); save result as a toast.
- [x] 12.3 Settings › Services also shows `rspamd.service`.
- [x] 12.4 Settings › Configuration: `GET/PUT /api/v1/system/config`, schema of editable/read-only/secret keys, line-preserving TOML editor (`utils.SetTOMLValue`, tested), validation, `.bak`, atomic write, audit; tab in PT/EN/ES at AAA; route tests (secrets hidden, 403, read-only/invalid refused, file updated in place).
- [x] 12.5 `terminal.enabled` editable in Settings › Configuration (confirmation to enable; open sessions closed within 15 s when disabled); host/port stay read-only.

## 13. Showcase README and polish

- [x] 13.1 Root README rewritten as a showcase (one-minute start, GHCR image, themes collage, feature table); advanced material moved to `DOCUMENTS/README.md` (install options, Docker image workflow and publishing, configuration, Settings, themes, reverse proxy, security, API, observability).
- [x] 13.2 Refreshed screenshots (current UI, soft shapes) plus `themes.png` and `settings-config.png`.
- [x] 13.3 Terminal tab reloads its settings when shown again (enabling in Configuration no longer needs F5).
- [x] 13.4 All dialogs open 5% below the top of the window.
- [x] 13.5 Menu icons on hover/active use `--color-icon-accent` (the theme's link color, >= 4:1 on the sidebar in every theme; Default unchanged): the action fill matched the Windows 98 sidebar (1:1) and was low on crystal-dark, graphite and grafana.
- [x] 13.7 Own toast module `frontend/src/lib/toast` (reactive queue, `ToastHost` with ARIA roles, pause on hover/focus loss, progress, reduced motion, tests, README) replaces vue3-toastify (dependency and vendored tarball removed); toasts mapped to theme roles.
- [x] 13.8 Cancel buttons (`btn-cancel`, 18 in dialogs): red text on hover/focus (`--color-danger-text`, >= 4.5:1 on the surface in every theme).
- [x] 13.9 Terminal errors (connection closed, auth, host key…) shown as toasts instead of a banner above the console, so the page no longer scrolls when a session ends.
- [x] 13.6 Text/focus uses of the action fill moved to `--color-icon-accent(-hover)` (links, row icons, focus borders, checkbox accent); menu hover tint/bar from the accent; `--color-accent-text` for accent badges; per-theme `--color-on-toast-*` (white/black by contrast, 76 pairs >= 4.5:1); EDIT buttons on the action fill in ported themes; literal `oklch()`/named colors removed and now rejected by the test. AA check of every non-default theme on dashboard, admins, mailboxes, domains, transports and API keys: no text below 4.5:1 (Default left as designed).

## 11. Final validation

- [x] 11.1 `openspec validate improve-ui-themes-and-platform --strict`.
- [x] 11.2 Settings at AAA (in-page checker) in default, xp, classic-gray, breeze-dark, matrix, high-contrast.
- [x] 11.3 `go test ./...`, `npm run check`, `npm run test`, `make build`, full visual matrix (20 themes × 7 screens × 4 sizes = 560 shots) green; modal text bug found by it (inherited CanvasText) fixed with `body { color: var(--color-fg) }`.
- [x] 11.4 Codex review of the full diff: 7 findings (arm64 TARGETARCH default, client IP behind proxies with Echo 5.4, host-key pin race, mailbox tokens on services, terminal reconnect after unmount, Vite WebSocket proxy, theme init race), all fixed with tests where testable; a verification pass confirmed them and the i18n, and found LAN clients could still spoof X-Forwarded-For — now only loopback and `server.trusted_proxies` are trusted (`evidence/codex-review-3-implementation.md`, `codex-review-4-verification.md`).
