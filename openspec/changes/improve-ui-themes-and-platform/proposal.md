## Why

The Vue SPA has a single hard-coded look: 520 hex colors spread over 33 `.vue`/`.css` files and only 4 `@media` queries. Operators cannot pick a theme, and on large monitors the layout does not scale: at 3840x2160 the login card is 372 px wide in the middle of an empty page (see `evidence/baseline-login-4k.png`). The platform is also behind: Go 1.26 (current 1.27.3), Echo v5.0.4 (current v5.4.4), `echo-contrib` v0.50.0 (retracted; Echo v5 must use the `echo-contrib/v5` module), plain Vite instead of Vite+ (`vp`, v1.5.1) used by the sibling project ad-manager, and the Docker image is not published anywhere — `README.md` documents `docker run` but users must build it themselves.

## What Changes

- Introduce a CSS role contract (semantic `--color-*` / recipe variables) in `frontend/src/style.css` and replace hard-coded colors in components with those roles. The current neo-brutalist look becomes the `default` theme with no visual change.
- Port the theme catalog from `jniltinho/ad-manager` (`frontend/src/themes/*.css`): default, admanager, breeze/breeze-dark, violet/violet-dark, aurora/aurora-dark, crystal/crystal-dark, cone, graphite, grafana, rabbitmq, matrix, winamp, pink-panther, high-contrast, classic-gray (Windows 98) and **XP** (`xp`, the Luna look ported from ad-manager's `classic-blue`).
- Add a **Settings** page (new sidebar entry — the old legacy-UI "Settings" link pointed to `#`) with tabs, like ad-manager: **Appearance** (theme cards with live previews), **Services** (read-only systemd status of Postfix and Dovecot, as in pf-report's Manager page) and **Terminal** (web SSH terminal, below). The mailbox portal gets a compact theme select in its header. Themes apply through `data-theme` + `color-scheme`, persist per browser in `localStorage` with an anti-flash inline script in `index.html`; the server default comes from a new `[ui] default_theme` config key.
- Add an opt-in web SSH terminal (Settings › Terminal) ported from ad-manager: xterm.js in the browser, an `xterm-256color` PTY over `golang.org/x/crypto/ssh`, carried by a WebSocket; superadmins only, credentials never stored, host key pinned on first use after the operator confirms its fingerprint, open/close audited, idle timeout. Its palette is a set of `--term-*` theme roles so every theme can style it.
- Settings meets WCAG AAA text contrast (7:1, 4.5:1 for large text) in every theme.
- New screens in PT, EN and ES; the language follows the browser and is chosen in Settings › Appearance (the PT/EN/ES buttons of the admin header and login pages are removed; the portal keeps a compact select).
- Make the layout scale for FullHD (1920x1080), QHD (2560x1440) and 4K (3840x2160): a `clamp()` root font size (16px up to FullHD — unchanged there — ~19px on QHD, 24px on 4K) with component sizes converted from `px` to `rem`, icons in `rem`, and wider login card and portal content on large screens.
- Migrate the frontend build to Vite+ (`vite-plus`, CLI `vp`): `vp dev`, `vp build`, `vp check`; keep `vue-tsc` type checking and the `web/dist` output used by `//go:embed`.
- Upgrade Go to 1.27 (`go.mod`, Dockerfile, CI), Echo to v5.4.x, replace `echo-contrib` with `github.com/labstack/echo-contrib/v5`, and patch-level updates of direct dependencies; fix any Echo v5 API drift.
- Publish a multi-arch (linux/amd64, linux/arm64) image to GitHub Packages (`ghcr.io/jniltinho/go-postfixadmin`) on every `v*` tag, with `latest`, semver tags and OCI labels.
- Rewrite `README.md` (themes gallery, GHCR quick start, supported resolutions) and update `DEVELOPMENT.md`, `FEATURES.md`, `CONTRIBUTING.md`, `AGENTS.md` and `DOCUMENTS/setup/README.md`.
- Validate visually with `agent-browser` (every theme x FullHD/QHD/4K) and run an independent review with Codex CLI before merge.

## Capabilities

### New Capabilities

- `ui-theming`: Role-based CSS theme contract, theme catalog including XP, selector, persistence and server default.
- `ui-high-resolution-layout`: Layout that scales across FullHD, QHD and 4K without empty space or tiny text.
- `build-toolchain`: Go 1.27, Echo v5.4.x and Vite+ (`vp`) as the supported build toolchain.
- `mail-service-status`: Settings › Services, read-only Postfix/Dovecot unit status.
- `web-ssh-terminal`: Settings › Terminal, the opt-in browser SSH terminal and its security rules.
- `container-image-publishing`: Multi-arch Docker image published to GHCR on release tags.

### Modified Capabilities

None. There are no archived OpenSpec capabilities in this repository yet.

## Impact

- Frontend: `frontend/src/style.css`, new `frontend/src/themes/*.css`, `frontend/index.html`, `frontend/src/layouts/MainLayout.vue`, `HeaderBar.vue`, `SidebarNav.vue`, `AppTable.vue`, login pages and the 33 files with hard-coded colors; new `frontend/src/stores/theme.ts`; `frontend/package.json`, `frontend/vite.config.ts`, lockfile.
- Backend: `go.mod`/`go.sum`; `[ui] default_theme` in `config.toml.example`, `web/files/config.default.toml`, exposed as an additive `default_theme` field on the public `GET /api/v1/version` response (no new route, no new permission).
- Build/CI: `Makefile` (`frontend` target uses `vp`), `Dockerfile` (`golang:1.27-alpine`, Vite+ in the Node stage), `.github/workflows/release.yml` (Go 1.27) and new `.github/workflows/docker.yml` (needs `packages: write`).
- i18n: theme labels and selector strings in `locales/{en,es,pt_BR}/default.po` and the SPA translations.
- Docs: README, DEVELOPMENT, FEATURES, CONTRIBUTING, AGENTS, setup guide, new screenshots.
- One new table `terminal_host_key` (GORM, in `MigrateDB`). No RBAC catalog change: Services reuses `settings:read`, Terminal requires superadmin (+ `settings:write` with RBAC on). API additions: `default_theme` on `/api/v1/version`, `GET /system/services`, `GET /system/terminal`, `DELETE /system/terminal/hostkey`, `GET /system/terminal/ws` (WebSocket). New dependencies: `golang.org/x/net/websocket` (already indirect), `@xterm/xterm`, `@xterm/addon-fit`.

## References

- Theme source: https://github.com/jniltinho/ad-manager/tree/main/frontend/src/themes (local: `/home/nilton/Projetos/nilton/NOVOS/ad-manager/frontend/src/themes`)
- Theme registry in ad-manager: `internal/server/themes.go`
- Vite+: https://viteplus.dev
- Echo v5 releases: https://github.com/labstack/echo/releases
- GHCR: https://docs.github.com/packages/working-with-a-github-packages-registry/working-with-the-container-registry
