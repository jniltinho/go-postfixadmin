# Go-PostfixAdmin

[![Go Version](https://img.shields.io/github/go-mod/go-version/jniltinho/go-postfixadmin)](https://go.dev/)
[![License](https://img.shields.io/github/license/jniltinho/go-postfixadmin)](./LICENSE)
[![Release](https://img.shields.io/github/v/release/jniltinho/go-postfixadmin?color=blue)](https://github.com/jniltinho/go-postfixadmin/releases)
[![Stars](https://img.shields.io/github/stars/jniltinho/go-postfixadmin?style=social)](https://github.com/jniltinho/go-postfixadmin/stargazers)
[![Docker Pulls](https://img.shields.io/docker/pulls/jniltinho/go-postfixadmin)](https://hub.docker.com/r/jniltinho/go-postfixadmin)

<p align="center">
  <img src="DOCUMENTS/screenshots/dashboard.png" alt="Go-PostfixAdmin dashboard" width="900">
</p>

<p align="center">
  <b>A modern, single-binary web panel and REST API for Postfix + Dovecot mail servers.</b><br>
  Domains, mailboxes, aliases, admins, RBAC and a self-service user portal — no PHP, no web server, no extra runtime.
</p>

<p align="center">
  <a href="#-quick-start">Quick Start</a> ·
  <a href="#-screenshots">Screenshots</a> ·
  <a href="#-features">Features</a> ·
  <a href="FEATURES.md">Full feature list</a> ·
  <a href="DOCUMENTS/setup/README.md">Setup guide</a>
</p>

---

## ✨ Why Go-PostfixAdmin?

- **One file to deploy** — the Vue 3 UI, Swagger docs and translations are embedded in a single static Go binary. Copy it, point it at your database, run it. Also shipped as Docker image, `.deb` and `.rpm`.
- **Fast and lightweight** — no PHP-FPM, Apache/Nginx or Node runtime in production — just one Go process.
- **API-first** — every action in the UI is available through a documented REST API (`/api/v1`) with JWT and long-lived API keys, ready for automation and provisioning scripts.
- **Built for teams** — superadmins, domain admins and fine-grained RBAC roles decide who can see and change what.
- **Users help themselves** — a separate portal lets mailbox owners change their password, set forwarding and configure vacation auto-replies.
- **Familiar data model** — PostfixAdmin-style tables (`domain`, `mailbox`, `alias`, …) and `$2y$` bcrypt hashes, on MariaDB/MySQL or PostgreSQL.
- **Scriptable CLI** — bulk CSV import/export, domain/mailbox/transport management, MySQL backups, quota reports and a Postfix `transport_maps` TCP server.

## 🚀 Quick Start

### Docker Compose (fastest)

```bash
git clone https://github.com/jniltinho/go-postfixadmin.git
cd go-postfixadmin
docker compose up -d
```

Open `http://localhost:8080` — the stack includes MariaDB and creates the first admin automatically (`admin@example.com` / `adminpassword` — change it in [docker-compose.yml](docker-compose.yml)).

### Docker

```bash
docker run -p 8080:8080 \
  -e DB_URL="postfix:pass@tcp(your-db:3306)/postfix?charset=utf8mb4&parseTime=True" \
  jniltinho/go-postfixadmin:latest
```

### Binary / .deb / .rpm

Download from [GitHub Releases](https://github.com/jniltinho/go-postfixadmin/releases), then:

```bash
./postfixadmin --generate-config      # writes config.toml — set your database
./postfixadmin migrate                # create / update the schema
./postfixadmin migrate rbac           # optional: RBAC roles and permissions
./postfixadmin admin --add-superadmin admin@example.com:'S3cure!Pass'
./postfixadmin server                 # http://localhost:8080
```

Swagger UI is available at `/swagger/` when `server.swagger_enable = true`.

### Build from source

```bash
make build        # builds the Vue 3 frontend and embeds it into bin/postfixadmin
make run
```

See [DEVELOPMENT.md](DEVELOPMENT.md) for the full developer workflow.

## 📸 Screenshots

### Administration

| Domains | Email accounts |
|---------|----------------|
| ![Domains](DOCUMENTS/screenshots/domains.png) | ![Mailboxes](DOCUMENTS/screenshots/mailboxes.png) |

| Add email account | Add domain |
|-------------------|------------|
| ![Add email account](DOCUMENTS/screenshots/create-mailbox.png) | ![Add domain](DOCUMENTS/screenshots/create-domain.png) |

| Aliases | Domain aliases |
|---------|----------------|
| ![Aliases](DOCUMENTS/screenshots/aliases.png) | ![Domain aliases](DOCUMENTS/screenshots/alias-domains.png) |

| Administrators | RBAC roles |
|----------------|------------|
| ![Administrators](DOCUMENTS/screenshots/admins.png) | ![Roles](DOCUMENTS/screenshots/roles.png) |

| Mail delivery log | Audit log |
|-------------------|-----------|
| ![Mail log](DOCUMENTS/screenshots/maillog.png) | ![Audit log](DOCUMENTS/screenshots/logs.png) |

### User self-service portal

| Forwarding & password | Vacation auto-reply |
|-----------------------|---------------------|
| ![User portal](DOCUMENTS/screenshots/user-dashboard.png) | ![Vacation](DOCUMENTS/screenshots/user-vacation.png) |

### REST API

![Swagger UI](DOCUMENTS/screenshots/swagger.png)

More screenshots (login pages, transports, API keys, create admin/alias…) in [DOCUMENTS/screenshots](DOCUMENTS/screenshots).

## 🧩 Features

- **Mail management** — domains (limits, quotas, backup MX), mailboxes with quota, aliases with multiple recipients, alias domains, per-domain Postfix transports
- **User portal** — password change, forwarding, vacation auto-reply via Dovecot Sieve
- **Security** — JWT (short-lived access token + httpOnly refresh cookie), CSRF origin check, login rate limiting, strong password policy with built-in generator, persistent API keys with expiry
- **Access control** — superadmin / domain-admin separation plus optional fine-grained RBAC roles
- **Operations** — admin audit log, mail delivery log ingestion (`readlog`), quota reports, MySQL/MariaDB backups, optional OpenTelemetry export to OpenObserve
- **i18n** — English, Português (pt_BR) and Español
- **Databases** — MariaDB/MySQL and PostgreSQL

Complete capability list: [FEATURES.md](FEATURES.md)

## Documentation

- [Features](FEATURES.md)
- [Development Guide](DEVELOPMENT.md)
- [Quick Setup – MariaDB](DOCUMENTS/setup/SETUP_MAILSERVER_MARIADB.md)
- [Quick Setup – PostgreSQL](DOCUMENTS/setup/SETUP_MAILSERVER_POSTGRESQL.md)
- [Complete Setup Guide](DOCUMENTS/setup/README.md)
- [Transport TCP Server](DOCUMENTS/setup/SETUP_TRANSPORT.md)
- [Contributing](CONTRIBUTING.md)

## Tech Stack

**Backend**: Go 1.26, Echo v5, GORM, Cobra, Viper, log/slog, Swaggo (OpenAPI)

**Frontend (embedded)**: Vue 3 + TS + Vite, Pinia, Vue Router, Tailwind CSS v4, Lucide icons

**Packaging**: Multi-stage Docker, UPX-compressed binary, native .deb/.rpm via Makefile + GitHub Actions

## Contributing

If Go-PostfixAdmin is useful to you, please ⭐ the repository — it helps other admins find it.


Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) first.

## License

Go-PostfixAdmin is an independent project, inspired by and not affiliated with [PostfixAdmin](https://github.com/postfixadmin/postfixadmin).


This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

## OpenTelemetry / OpenObserve

Optional application log, HTTP/database trace and metric export to OpenObserve is available. All export and database tracing switches default to disabled. See [configuration and deployment](DOCUMENTS/setup/OBSERVABILITY.md) and [isolated Docker validation](tests/observability/README.md).

Set the OpenObserve authorization header using your ingestion email/password (the credentials below are fictional):

```bash
# Fictional credentials: replace with your OpenObserve ingestion credentials.
export OBSERVABILITY_AUTHORIZATION="Basic $(printf '%s' 'telemetry@example.com:example-password' | base64 | tr -d '\r\n')"
```

Then enable the desired signals and set their endpoints as described in the configuration guide. For Docker, pass this variable into the app service and recreate the container.
