# Contributing to Go-Postfixadmin

First off, thank you for considering contributing to Go-Postfixadmin! It's people like you that make this a great tool for the community.

## 🚀 How Can I Contribute?

### Reporting Bugs
If you find a bug, please open an issue on GitHub. Include:
- A clear and descriptive title.
- Steps to reproduce the bug.
- Actual vs. expected behavior.
- Screenshots if applicable.

### Suggesting Enhancements
Have an idea to make Go-Postfixadmin better?
- Open an issue with the "enhancement" label.
- Describe the feature and why it would be useful.

### Pull Requests
1. **Fork the repository** and create your branch from `main`.
2. **Setup your environment**:
   - Install Go (v1.22+) and Node.js (v20+).
   - Run `make deps` to install all Go and NPM dependencies.
3. **Make your changes**:
   - Follow clean code principles: concise, self-documenting, no over-engineering.
   - If adding a new feature, ensure it's documented.
   - If adding a new language, add a `.po` file to `locales/` (see [Localization](#-localization-i18n) below).
4. **Test your changes**:
   - Run `make run` to build and start the server locally.
   - Ensure the UI looks good across different screen sizes.
5. **Submit your PR**:
   - Provide a concise title and detailed description of your changes.
   - Reference any related issues.

---

## 🏗 Project Structure

```
.
├── cmd/                  # Cobra CLI commands (root, server, admin, version, migrate-import, config-generator)
├── admin/                # Admin CLI logic: listing, cleanup, quota report
├── internal/
│   ├── handlers/         # HTTP route handlers
│   ├── i18n/             # Internationalization (gotext wrapper)
│   ├── middleware/        # HTTP middleware (auth, logging, etc.)
│   ├── models/           # GORM database models
│   ├── routes/           # Route definitions
│   ├── server/           # HTTP server setup and render helpers
│   └── utils/            # Shared utilities (mailer, DB connection, quota)
├── locales/              # GNU Gettext .po translation files
├── web/
│   ├── static/           # Static assets (CSS, JS, images)
│   └── templates/        # HTML templates (Go html/template)
├── config.toml.example   # Example configuration file
└── Makefile              # Build and development commands
```

---

## 🖥 CLI Reference

The binary is named `postfixadmin` and uses [Cobra](https://github.com/spf13/cobra) for its CLI. Available subcommands:

| Command | Description |
|---|---|
| `postfixadmin server` | Start the web server |
| `postfixadmin admin` | Admin management utilities (see flags below) |
| `postfixadmin version` | Print the current version |
| `postfixadmin migrate-import` | Import data from a legacy PostfixAdmin database |
| `postfixadmin --generate-config` | Generate a default `config.toml` in the current directory |

### Global Flags

| Flag | Description |
|---|---|
| `--config <path>` | Path to config file (default: `./config.toml`, `/etc/postfixadmin/config.toml`) |
| `--db-url <url>` | Database connection string (overrides `config.toml`) |
| `--db-driver <driver>` | Database driver: `mysql` or `postgres` |

> **Tip:** You can also set `DATABASE_URL` and `DATABASE_DRIVER` as environment variables — Viper picks them up automatically via `AutomaticEnv()`.

### `admin` Subcommand Flags

| Flag | Short | Description |
|---|---|---|
| `--list-domains` | `-d` | List all domains |
| `--list-mailboxes` | `-m` | List all mailboxes |
| `--list-admins` | `-a` | List all administrators |
| `--list-aliases` | `-s` | List all aliases |
| `--list-alias-domains` | `-S` | List all alias domains |
| `--domain-admins` | `-A` | List all domain admins |
| `--list-logs` | `-L` | List the last 100 system log entries |
| `--quota-report` | `-q` | Show Dovecot quota report (requires `doveadm`) |
| `--email <addr>` | `-e` | Send the quota report via SMTP to this address |
| `--cleanup-maildir` | `-c` | Clean up orphaned maildirs on disk |
| `--base-dir <path>` | | Base directory for maildirs (default: `/var/vmail`) |
| `--add-superadmin <email:pass>` | | Create a new superadmin user |

All tabular output uses [`go-pretty`](https://github.com/jedib0t/go-pretty) for clean, formatted console tables.

---

## ⚙️ Configuration

Copy `config.toml.example` to `config.toml` and adjust as needed. Key sections:

```toml
[database]
url    = "user:password@tcp(host:3306)/dbname?parseTime=True"
# driver defaults to "mysql"; use "postgres" for PostgreSQL

[server]
port            = 8080
cleanup_maildir = false   # Auto-clean orphaned maildirs on mailbox deletion

[ssl]
session_secret = "generate-with: openssl rand -hex 32"

[quota]
enabled      = false
domain_quota = true
multiplier   = 1048576   # 1 MB = 1048576 bytes

[smtp]
server = "localhost"
port   = 25
type   = "plain"   # Options: plain | tls | starttls
```

---

## 📧 SMTP & Welcome Emails

Welcome emails are sent when a new mailbox is created. The email subject and body are **automatically translated** based on the user's interface language. Supported languages:

| Code | Language |
|---|---|
| `en` | English |
| `pt-br` | Brazilian Portuguese |
| `es` | Spanish |

The SMTP connection supports three modes (`smtp.type` in `config.toml`):
- `plain` — Direct SMTP without TLS (suitable for local submission on port 25).
- `tls` — SMTP over TLS (e.g., port 465).
- `starttls` — SMTP with STARTTLS upgrade (e.g., port 587).

---

## 🌐 Localization (i18n)

The project uses [gotext](https://github.com/leonelquinteros/gotext) with GNU Gettext `.po` files.

To add a new language:
1. Create a new directory in `locales/` (e.g., `locales/fr/`).
2. Copy `locales/en/default.po` to `locales/fr/default.po`.
3. Translate the `msgstr` values (keep `msgid` keys unchanged).
4. Add a language switcher link in:
   - `web/templates/layout.html`
   - `web/templates/login.html`
   - `web/templates/users/layout.html`
   - `web/templates/users/login.html`
5. Register the new language code in `SetLanguage` (`internal/handlers/handlers.go`) and `Render` (`internal/server/render.go`).
6. Add translations for welcome email subject/body in `internal/utils/mailer.go` (the `SendWelcomeEmail` function uses `i18n.Translate`).

---

## 🎨 Design Guidelines

We use a **Neo-Brutalism** design aesthetic:
- Thick black borders (`2px` or `4px`).
- High-contrast colors.
- Sharp shadows (`neo-shadow`).
- Lucide icons throughout the UI.
- CSS built with **Tailwind CSS** (see `tailwind.config.js`).

---

## 🛠 Useful Commands

| Command | Description |
|---|---|
| `make run` | Build (with CSS) and start the server (`./postfixadmin server`) |
| `make build` | Compile the binary (with CSS, no UPX) |
| `make build-prod` | Compile and compress with UPX for production |
| `make css` | Build Tailwind CSS (minified) |
| `make watch-css` | Watch for CSS changes and rebuild automatically |
| `make deps` | Install Go modules and NPM packages |
| `make tidy` | Run `go mod tidy` |
| `make certs` | Generate self-signed SSL certificates |
| `make clean` | Remove the binary and generated CSS |
| `make build-docker` | Build the Docker image |

---

## 📜 Code of Conduct

Please be respectful and professional in all interactions within the project.

---

Happy coding! 🚀
