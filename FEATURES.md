# Features

Go-PostfixAdmin exposes a full set of capabilities for Postfix + Dovecot mail server administration via web UI, REST API (`/api/v1`), and CLI. 

For installation see the [setup guides](DOCUMENTS/setup/README.md) and [quick MariaDB summary](DOCUMENTS/setup/SETUP_MAILSERVER_MARIADB.md). For development and build, see [DEVELOPMENT.md](DEVELOPMENT.md). Overview and quick start: [README.md](README.md).

## Core

- Domain management
- Mailbox management
- Alias management
- Alias domain management
- Admin management
- Transport list management (Postfix transport routing per domain)
- User self-service portal

## REST API

A fully documented REST API at `/api/v1` with interactive Swagger UI at `/swagger/`.

### Authentication

- JWT-based: `POST /api/v1/auth/login` issues a short-lived access token (`Authorization: Bearer <token>`) and a long-lived httpOnly refresh cookie
- Token refresh: `POST /api/v1/auth/refresh`
- API Keys: persistent tokens for external integrations and automation (`/api/v1/settings/apikeys`)

### Endpoints

| Tag | Routes |
|-----|--------|
| Authentication | `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout`, `GET /auth/me` |
| Domains | `GET/POST /domains`, `GET/PUT/DELETE /domains/{domain}` |
| Mailboxes | `GET/POST /mailboxes`, `GET/PUT/DELETE /mailboxes/{username}` |
| Aliases | `GET/POST /aliases`, `GET/PUT/DELETE /aliases/{address}` |
| Alias Domains | `GET/POST /alias-domains`, `GET/PUT/DELETE /alias-domains/{alias_domain}` |
| Admins | `GET/POST /admins`, `GET/PUT/DELETE /admins/{username}` |
| Transports | `GET/POST /transports`, `GET/PUT/DELETE /transports/{id}` |
| User Portal | `GET /user/me`, `GET/POST /user/forwarding`, `POST /user/password`, `GET/POST/DELETE /user/vacation` |
| Logs | `GET /logs` (admin action logs), `GET /maillog` (SMTP logs) |
| Dashboard | `GET /dashboard` |
| Settings | `GET/POST /settings/apikeys`, `PUT/DELETE /settings/apikeys/{id}` |

All endpoints return a structured `APIResponse` envelope with `success`, `data`, and `error` fields. All request and response schemas are typed and visible in the Swagger UI.

## Access Control

- Superadmin and domain-admin separation via RBAC (enable with `rbac.enabled = true` after running `postfixadmin migrate rbac`)
- `rbac` CLI subcommand for assigning roles (`rbac assign`, `rbac seed-existing`)
- User self-service portal for password, forwarding, and vacation management (no admin rights needed)
- JWT access tokens (short-lived, `jwt_access_ttl`) + httpOnly refresh cookies (`jwt_refresh_ttl`)
- Persistent API keys (with optional expiry) under `/api/v1/settings/apikeys`

## Passwords

- Shared backend password policy for mailbox, admin, and user portal password changes
- Minimum 8 characters
- At least 1 uppercase letter
- At least 1 lowercase letter
- At least 1 number
- At least 1 special character
- Frontend password generator compatible with backend validation
- Bcrypt hashing with PostfixAdmin-compatible `$2y$` prefix

## Mail Features

- Optional welcome email on mailbox creation
- Vacation / auto-reply integration with Dovecot Sieve
- Forwarding management in the user portal
- Quota-aware mailbox model support

## Logs and Operations

- Admin log viewer with filtering, pagination, and free-text search
- Mail log viewer with filtering and pagination
- `readlog` daemon to ingest `FILTER:` entries from `/var/log/mail.log`
- CLI utilities for migration, SQL import, admin recovery, and maintenance tasks
- CLI `backup-mysql` subcommand:
  - List all MySQL/MariaDB databases with sizes (`backup-mysql list`)
  - Backup all non-system databases as compressed `.sql.gz` files (`backup-mysql backup`)
  - Automatic cleanup of backup files older than N days (`--clean N`)
  - Optional log delivery by e-mail via SMTP (`--sendmail`)
  - Verbose mode that prints log output to stdout (`--verbose`)
  - Config via `config.toml [backup]` section, environment variables, or CLI flags (`--host`, `--user`, `--passwd`)
- CLI `mailbox` subcommand:
  - List all mailboxes
  - Create individual mailbox with configurable quota (`--quota`)
  - Bulk-import from CSV (`--import-csv`), with support for pre-hashed passwords (`--password-crypt`)
  - Export all mailboxes to CSV (`--export`), compatible with re-import for backup and migration workflows
- CLI `domain` subcommand:
  - List all domains
  - Create a domain with optional description, alias and mailbox limits (`--add`, `--description`, `--max-aliases`, `--max-mailboxes`)
  - Delete a domain and all its associated data (mailboxes, aliases, alias domains, fetchmail, vacation) (`--delete`)
- CLI `transport` subcommand:
  - List all transport entries
  - Create a transport entry (`--add "domain:transport"`)
  - Delete a transport entry by domain (`--delete`)
  - `transport server`: TCP server for Postfix `transport_maps` lookups (`transport_maps = tcp:127.0.0.1:12221`)
    - Single-query JOIN resolves per-user, per-domain, and domain-default transport in priority order
    - In-memory cache with configurable TTL (`transport.cache`)
    - Colored zerolog console output with `--debug` flag for per-request tracing (source, subject, result)
    - Config via `config.toml [transport]` section (`host`, `cache`, `hostname`, `localdelivery`, `delivery`)
    - systemd service file included (`postfixadmin-transport.service`)

## Internationalization

- GNU Gettext `.po`-based i18n
- Portuguese (`pt_BR`), English (`en`), and Spanish (`es`)
- Backend flash messages translated through the same i18n layer
- All management UIs fully translated, including the transport list CRUD

## UI / DX

- Vue 3 SPA (Vite, Pinia, Vue Router, Tailwind CSS v4) with frontend fully embedded in the Go binary via `//go:embed` — single deployable artifact
- Interactive Swagger UI at `/swagger/` (toggle with `server.swagger_enable = true`)
- Docker, Docker Compose, Makefile, native build, plus `.deb` / `.rpm` packaging (`make deb`, `make rpm`) with config embedding
- Fully automated releases via GitHub Actions (`.tar.gz`, `.deb`, `.rpm`)

## Configuration Highlights

Key sections in `config.toml` (or env / CLI overrides):

- `[database]` — MariaDB/MySQL or PostgreSQL (driver, host, etc.); `debug = true` for verbose GORM SQL
- `[server]` — port, `ssl_enable`, cert/key paths, `swagger_enable`, JWT TTLs (`jwt_access_ttl`, `jwt_refresh_ttl`), `cleanup_maildir`
- `[quota]`, `[vacation]`, `[alias]`, `[transport]`, `[features]` (e.g. `fetchmail`), `[rbac]` (`enabled = true` after `migrate rbac`), `[smtp]` (for welcome emails), `[backup]`
- SSL-ready, session secret, API key support

See `config.toml.example` and the setup guides for full details.
