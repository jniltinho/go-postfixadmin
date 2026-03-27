# Features

## What This Document Covers

This file lists product capabilities by area. For installation and deployment details, use the setup guides:

- [Project README](README.md)
- [Quick mail server setup](DOCUMENTS/setup/SETUP_MAILSERVER.md)
- [Complete setup guide](DOCUMENTS/setup/README.md)

## Core

- Domain management
- Mailbox management
- Alias management
- Alias domain management
- Admin management
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
- CLI `mailbox` subcommand: list, create (with configurable quota), and bulk-import mailbox users from CSV

## Internationalization

- GNU Gettext `.po`-based i18n
- Portuguese (`pt_BR`), English (`en`), and Spanish (`es`)
- Backend flash messages translated through the same i18n layer

## UI / DX

- Tailwind CSS interface
- Modal-based admin workflows
- AJAX-based user password and forwarding updates
- Docker, Docker Compose, Makefile, and native build support

## Configuration Highlights

- MariaDB and PostgreSQL support
- `database.debug = true` for verbose GORM SQL logging
- SSL-ready server configuration
- Vacation, alias, quota, transport, and fetchmail feature flags
