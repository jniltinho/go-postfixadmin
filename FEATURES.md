# Features

## What This Document Covers

This file lists product capabilities by area. For installation and deployment details, use the setup guides:

- [Project README](README.md)
- [Quick mail server setup](DOCUMENTS/setup/SETUP_MAILSERVER_MARIADB.md)
- [Complete setup guide](DOCUMENTS/setup/README.md)

## Core

- Domain management
- Mailbox management
- Alias management
- Alias domain management
- Admin management
- Transport list management (Postfix transport routing per domain)
- User self-service portal

## Access Control

- Superadmin and domain admin separation
- User session area for password, forwarding, and vacation management
- Session-based authentication with inactivity timeout

## Passwords

- Shared backend password policy for mailbox, admin, and user password changes
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

- Admin log viewer
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

- Tailwind CSS interface
- Modal-based admin workflows
- AJAX-based user password and forwarding updates
- Docker, Docker Compose, Makefile, and native build support
- Debian (`.deb`) and RPM (`.rpm`) package build targets with automatic configuration embedding (`make deb` and `make rpm`)
- Fully automated CI/CD pipeline to generate `.tar.gz`, `.deb` and `.rpm` packages via GitHub Actions Releases

## Configuration Highlights

- MariaDB and PostgreSQL support
- `database.debug = true` for verbose GORM SQL logging
- SSL-ready server configuration
- Vacation, alias, quota, transport, and fetchmail feature flags
