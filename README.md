# Go-PostfixAdmin

Professional Email Administration System built with Go, Echo v5, Vue 3, and Tailwind CSS v4.

## What This Document Covers

This is the main entry point for the project documentation. It covers the product overview, build options, runtime basics, and links to the detailed guides.

## Overview

Go-PostfixAdmin provides a modern web UI, a full REST API, and CLI utilities for managing domains, mailboxes, aliases, admins, vacation rules, and operational mail logs in Postfix/Dovecot environments.

The frontend is a Vue 3 single-page application (SPA) embedded directly in the Go binary. The backend exposes a versioned REST API (`/api/v1`) with JWT authentication and interactive Swagger documentation available at `/swagger/`.

For the full feature list and capability breakdown, see [FEATURES.md](FEATURES.md).

## Documentation

- [Features](FEATURES.md)
- [Development guide](DEVELOPMENT.md)
- [Quick mail server setup](DOCUMENTS/setup/SETUP_MAILSERVER_MARIADB.md)
- [Complete setup guide](DOCUMENTS/setup/README.md)
- [Contributing](CONTRIBUTING.md)

---

## Execution

After building, run the binary directly:

```bash
./postfixadmin server --port=8080
```

Or via Docker:

```bash
make build-docker
docker run -p 8080:8080 -e DB_URL="your-dsn" jniltinho/postfixadmin:latest
```

The admin UI is available at `http://localhost:8080/`.
The Swagger UI is available at `http://localhost:8080/swagger/` (when `swagger_enable = true` in config).

---

## Screenshots

| Login | Admin Dashboard |
|-------|-----------------|
| ![Login](DOCUMENTS/screenshots/admin-login.png) | ![Dashboard](DOCUMENTS/screenshots/admin-welcome.png) |

| Domain List | Mailboxes |
|-------------|-----------|
| ![Domains](DOCUMENTS/screenshots/admin-domain-list.png) | ![Mailboxes](DOCUMENTS/screenshots/mailboxes-and-forwards-for-domain.png) |

| Alias Creation | Admin List |
|----------------|-----------|
| ![Alias](DOCUMENTS/screenshots/create-new-alias.png) | ![Admins](DOCUMENTS/screenshots/admin-list.png) |

| User Portal — Vacation | User Portal — Forwarding |
|------------------------|--------------------------|
| ![Vacation](DOCUMENTS/screenshots/users-enable-vacation-autoresponse.png) | ![Forwarding](DOCUMENTS/screenshots/users-edit-mail-forward.png) |

More images in the [screenshots folder](DOCUMENTS/screenshots).
