# Go-PostfixAdmin documentation

The [project README](../README.md) is the short tour. This guide covers installation options, the Docker image and
how it is published, configuration, the Settings page, themes, security and integrations.

| Topic | Where |
|-------|-------|
| Full mail server setup (Postfix, Dovecot, MariaDB, Sieve, rsyslog) | [setup/README.md](setup/README.md) |
| Short setups | [MariaDB](setup/SETUP_MAILSERVER_MARIADB.md) · [PostgreSQL](setup/SETUP_MAILSERVER_POSTGRESQL.md) |
| Postfix `transport_maps` TCP server | [setup/SETUP_TRANSPORT.md](setup/SETUP_TRANSPORT.md) |
| OpenTelemetry / OpenObserve | [setup/OBSERVABILITY.md](setup/OBSERVABILITY.md) |
| Every feature and API route | [FEATURES.md](../FEATURES.md) |
| Building, CLI reference, frontend workflow | [DEVELOPMENT.md](../DEVELOPMENT.md) |
| Contributing, adding a theme or a language | [CONTRIBUTING.md](../CONTRIBUTING.md) |
| Screenshot gallery | [screenshots/README.md](screenshots/README.md) |

**Contents:** [Install](#install) · [Docker image](#docker-image) · [Configuration](#configuration) ·
[Settings](#settings) · [Themes and layout](#themes-and-layout) · [Reverse proxy and TLS](#reverse-proxy-and-tls) ·
[Security](#security) · [API and automation](#api-and-automation) · [Observability](#observability) ·
[Quality checks](#quality-checks)

---

## Install

### Docker Compose

`docker-compose.yml` starts MariaDB and the panel from the published image, runs the migrations and creates the
first superadmin from `ADMIN_EMAIL` / `ADMIN_PASSWORD`:

```bash
git clone https://github.com/jniltinho/go-postfixadmin.git
cd go-postfixadmin
docker compose up -d        # http://localhost:8080
```

To build from your checkout instead, uncomment `build: .` in the `postfixadmin` service.

### Docker

```bash
docker run -d --name postfixadmin -p 8080:8080 \
  -e DB_DRIVER=mysql \
  -e DB_URL="postfix:pass@tcp(db:3306)/postfix?charset=utf8mb4&parseTime=True&loc=Local" \
  -e ADMIN_EMAIL=admin@example.com -e ADMIN_PASSWORD='change-me' \
  -v "$PWD/config.toml:/app/config.toml" \
  ghcr.io/jniltinho/go-postfixadmin:latest
```

| Variable | Use |
|----------|-----|
| `DB_URL`, `DB_DRIVER` | Database DSN and `mysql` / `postgres` (override `config.toml`) |
| `DB_HOST`, `DB_PORT` | If set, the entrypoint waits for the database before starting |
| `ADMIN_EMAIL`, `ADMIN_PASSWORD` | First superadmin, created on start if missing |
| `UI_DEFAULT_THEME` | Default theme for browsers without a saved choice |
| `OBSERVABILITY_*` | Telemetry settings, see [OBSERVABILITY.md](setup/OBSERVABILITY.md) |

The entrypoint runs `postfixadmin migrate` on every start (it only adds tables and columns).

### Binary, .deb, .rpm

Download from [Releases](https://github.com/jniltinho/go-postfixadmin/releases), then:

```bash
./postfixadmin --generate-config                 # writes config.toml with every key and its default
./postfixadmin migrate                           # core schema (safe to run on every upgrade)
./postfixadmin migrate rbac                      # optional: RBAC tables and built-in roles
./postfixadmin admin --add-superadmin admin@example.com:'S3cure!Pass'
./postfixadmin server                            # http://localhost:8080
```

systemd units for the panel and the transport server are in [setup/](setup/postfixadmin.service).

### From source

Go 1.27+, Node.js 22.18+ (or 24.11+) and make:

```bash
make build          # Vue 3 SPA (Vite+) embedded into bin/postfixadmin
make build-prod     # static binary compressed with UPX
make build-docker   # local image
make deb && make rpm
```

---

## Docker image

Images are built by [`.github/workflows/docker.yml`](../.github/workflows/docker.yml) and published to GitHub
Packages: **`ghcr.io/jniltinho/go-postfixadmin`**.

| | |
|---|---|
| **Platforms** | `linux/amd64`, `linux/arm64` (one multi-arch manifest) |
| **Triggers** | a pushed tag `v*.*.*`, or a manual run (*Actions › Docker image › Run workflow*) |
| **Tags** | `1.2.3` and `1.2` for every release tag; `latest` only for stable tags (no `-rc`, `-beta`…); `sha-<commit>` for manual runs |
| **Base** | `alpine:3.21` with `ca-certificates` and `tzdata`; the binary is static and UPX-compressed |
| **Metadata** | OCI labels and index annotation: source repository, description, license, version, revision |

How it is built: the frontend and Go stages run on the build machine (`--platform=$BUILDPLATFORM`) and Go
cross-compiles for each target (`GOOS/GOARCH` from buildx), so arm64 does not need emulation for the slow part;
only the final Alpine stage runs per platform. Layers are cached in GitHub Actions (`type=gha`).

**Publishing a release** (binaries, `.deb`, `.rpm` and the image run in parallel):

```bash
git tag v1.0.89
git push origin v1.0.89
```

**Manual image** from any branch (pushed as `sha-<commit>`, `latest` untouched):

```bash
gh workflow run docker.yml --ref main
```

**First publication:** GitHub creates the package as private. Open *Packages › go-postfixadmin › Package settings*,
set the visibility to **Public** and link it to the repository, so `docker pull` works without logging in.

---

## Configuration

`config.toml` is read from `--config`, `./config.toml`, `/etc/postfixadmin/` or `~/.postfixadmin/`.
`postfixadmin --generate-config` writes a complete file; [config.toml.example](../config.toml.example) documents every
key. Precedence: command-line flags › environment variables › config file › defaults.

| Section | Main keys |
|---------|-----------|
| `[database]` | `driver` (`mysql`/`postgres`), `host`, `port`, `user`, `pass`, `name`, `debug` |
| `[server]` | `port`, `ssl_enable`, `ssl_cert`, `ssl_key`, `session_secret`, `jwt_secret`, `jwt_access_ttl`, `jwt_refresh_ttl`, `trusted_proxies`, `swagger_enable`, `cleanup_maildir` |
| `[ui]` | `default_theme` |
| `[terminal]` | `enabled` (default `false`), `host`, `port`, `idle_timeout` |
| `[rbac]` | `enabled` — after `migrate rbac` and assigning roles |
| `[smtp]` | `server`, `port`, `type` (`plain`/`tls`/`starttls`) for welcome e-mails |
| `[vacation]`, `[quota]`, `[transport]`, `[backup]`, `[observability]` | see the example file |

Some keys can also be changed from the panel, see [Settings › Configuration](#configuration-tab).

---

## Settings

The **Settings** page (sidebar) has four tabs.

### Appearance

Language (Automatic = browser, English, Português, Español) and theme, both saved per browser. The theme cards
are live previews. `[ui] default_theme` applies to browsers without a saved choice.

### Services

Read-only systemd state of `postfix.service`, `dovecot.service` and `rspamd.service` (`systemctl show`: active state,
boot state, since). Needs systemd on the host where the panel runs; in a container it says systemd is unavailable.
RBAC permission: `settings:read`.

### Configuration tab

Superadmins only (with RBAC also `settings:read` / `settings:write`).

| Editable from the panel | Read-only (change on the server) |
|-------------------------|----------------------------------|
| `ui.default_theme`, `vacation.enabled`, `smtp.server`, `smtp.port`, `smtp.type`, `server.swagger_enable` (after a restart), `terminal.enabled`, `terminal.idle_timeout` | database, server port/TLS, secrets, JWT lifetimes, `trusted_proxies`, `cleanup_maildir`, quota multiplier, RBAC, transport, terminal host/port, backup, observability |

Secrets are never sent to the browser (only "set" / "not set"). Saving validates each value, changes only those lines
of `config.toml` (comments and alignment kept), checks the new file parses, keeps the previous one as
`config.toml.bak`, replaces the file atomically with its permissions and logs `config_update` in the audit log.
A key set by an environment variable, or a file the panel cannot write, is shown locked.

### Terminal

A web SSH terminal (xterm.js, `xterm-256color` PTY, 256 colors, classic xterm palette) to `terminal.host:port`:

- **off by default**; a superadmin enables it in the Configuration tab (with a confirmation) or in `config.toml`;
  disabling closes open sessions within 15 seconds;
- superadmins only, checked by the server; the WebSocket accepts only the panel's own origin and gets the token in
  its first message (never in the URL);
- the SSH password or key is typed in the browser, used to sign in and dropped — never stored;
- first connection: the panel shows the host key fingerprint; compare it with
  `ssh-keygen -lf /etc/ssh/ssh_host_ed25519_key.pub` on the server. It is then pinned (`terminal_host_key` table)
  and a different key is refused until a superadmin uses *Forget host key*;
- opening, failing, closing (with duration) and trusting a key are written to the audit log; the commands typed
  are not — the server's own login records are the record;
- at most 4 terminals per admin; closed after `idle_timeout` without typing.

---

## Themes and layout

20 themes, ported from [ad-manager](https://github.com/jniltinho/ad-manager): Default (neo-brutalist), ADManager,
Breeze, Violet, Aurora and Crystal (light and dark), Cone, Windows XP, Windows 98, Graphite, Grafana, RabbitMQ,
Matrix, Winamp, Pink Panther and High contrast.

- Colors, corners and shadows are **roles** (`--color-*`, `--radius`, `--shadow-N`). Default keeps square corners,
  hard offset shadows and strong lines; the other themes get subtle rounded corners, soft shadows and lighter lines
  (High contrast keeps strong lines).
- The UI is sized in `rem` on a root font of 16px up to FullHD, ~19px on QHD and 24px on 4K, so it scales with the
  screen; at FullHD it renders exactly as designed.
- Settings meets WCAG AAA text contrast in every theme checked.
- Adding a theme: [CONTRIBUTING.md › Themes](../CONTRIBUTING.md#-themes).

---

## Reverse proxy and TLS

The panel can serve HTTPS itself (`ssl_enable = true`, `ssl_cert`, `ssl_key`) or sit behind a proxy. Behind nginx:

```nginx
location / {
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header Host $host;                     # the terminal checks Origin against Host
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_http_version 1.1;                          # WebSocket for Settings › Terminal
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
}
```

The login rate limiter and the audit log use the client IP. `X-Forwarded-For` is trusted only from loopback (a
proxy on the same host) and from `[server] trusted_proxies`, e.g. `["172.18.0.0/16"]` for a proxy in a Docker
network; any other client, LAN ones included, is identified by its socket address and cannot spoof it.

---

## Security

- **Sessions**: short-lived JWT access token (`jwt_access_ttl`, 15m) plus an httpOnly refresh cookie
  (`jwt_refresh_ttl`, 7 days).
- **Login**: rate limited per IP (3 attempts, then 1 every 12 s); CSRF origin check on auth endpoints.
- **Access**: superadmins, domain admins limited to their domains, and optional RBAC roles and permissions
  (`migrate rbac`, then `[rbac] enabled = true`).
- **API keys**: long-lived, per admin, with expiry, sent as `X-API-Key` or `Authorization: Bearer`.
- **Passwords**: strong-password policy and generator; `$2y$` bcrypt hashes compatible with PostfixAdmin.
- **Settings**: secrets never leave the server; terminal off by default and audited.

---

## API and automation

Everything in the UI is available under `/api/v1` (JSON envelope `{success, data, error}`). Set
`server.swagger_enable = true` for the interactive docs at `/swagger/`.

```bash
TOKEN=$(curl -s -X POST https://mail.example.com:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin@example.com","password":"..."}' | jq -r .data.access_token)
curl -s -H "Authorization: Bearer $TOKEN" https://mail.example.com:8080/api/v1/domains | jq
```

For scripts, create an API key in *API Keys* and send it as `X-API-Key`. The CLI covers bulk work too: CSV import
and export of mailboxes, domains, transports, MySQL backups and quota reports — see
[DEVELOPMENT.md › CLI](../DEVELOPMENT.md#cli-overview).

---

## Observability

Optional export of application logs, HTTP/database traces and metrics over OTLP/HTTP to OpenObserve. Every switch
is off by default and local logging always stays on. Configuration and limits: [setup/OBSERVABILITY.md](setup/OBSERVABILITY.md);
isolated Docker validation: [tests/observability](../tests/observability/README.md).

```bash
# fictional credentials: use your OpenObserve ingestion user
export OBSERVABILITY_AUTHORIZATION="Basic $(printf '%s' 'telemetry@example.com:example-password' | base64 | tr -d '\r\n')"
```

---

## Quality checks

| Check | Command |
|-------|---------|
| Go tests | `go test ./...` |
| Frontend format, lint, unit tests | `cd frontend && npm run check && npm run test` |
| Visual matrix (themes × pages × 1366/FHD/QHD/4K) | [tests/visual/capture.sh](../tests/visual/README.md) |
| API docs | `make swagger` |
