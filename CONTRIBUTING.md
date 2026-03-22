# Contributing to Go-PostfixAdmin

Thank you for contributing to Go-PostfixAdmin.

This guide covers the contribution workflow, project conventions, and the places where implementation and documentation changes usually need to stay aligned.

Related documents:

- [Project README](README.md)
- [Development guide](DEVELOPMENT.md)
- [Features](FEATURES.md)
- [Quick mail server setup](DOCUMENTS/setup/SETUP_MAILSERVER.md)
- [Complete setup guide](DOCUMENTS/setup/README.md)

## 🚀 How Can I Contribute?

### Reporting Bugs
If you find a bug, please open an issue on GitHub. Include:
- A clear and descriptive title.
- Steps to reproduce the bug.
- Actual vs. expected behavior.
- Screenshots if applicable.

### Suggesting Enhancements
Have an idea to make Go-PostfixAdmin better?
- Open an issue with the "enhancement" label.
- Describe the feature and why it would be useful.

### Pull Requests
1. **Fork the repository** and create your branch from `main`.
2. **Setup your environment**:
   - Install Go (v1.22+) and Node.js (v20+).
   - Run `make deps` to install all Go and NPM dependencies.
3. **Make your changes**:
   - Follow clean code principles: concise, self-documenting, no over-engineering.
   - If adding a new feature, ensure it is documented in the appropriate file.
   - If adding a new language, add a `.po` file to `locales/` (see [Localization](#-localization-i18n) below).
4. **Test your changes**:
   - Run `go test ./...`.
   - Run `make run` when the change affects the running application or UI.
   - Run `make build-prod` when your change affects embedded assets, build flow, or release packaging.
   - Ensure the UI looks good across different screen sizes.
5. **Submit your PR**:
   - Provide a concise title and detailed description of your changes.
   - Reference any related issues.

### Documentation Expectations

Update documentation when your change affects:

- user-facing behavior
- configuration keys or defaults
- CLI commands or flags
- setup or deployment steps
- password policy or authentication flow
- translations or language support

Typical locations:

- `README.md`: project overview and navigation
- `DEVELOPMENT.md`: local build, CLI, and development workflow
- `FEATURES.md`: capability overview
- `DOCUMENTS/setup/SETUP_MAILSERVER.md`: quick setup summary
- `DOCUMENTS/setup/README.md`: full setup and deployment guide

---

## 🏗 Project Structure

```
.
├── cmd/                  # Cobra CLI commands (root, server, admin, version, migrate, importsql, config-generator)
├── admin/                # Admin CLI logic: listing, cleanup, quota report
├── internal/
│   ├── handlers/         # HTTP route handlers
│   ├── i18n/             # Internationalization (gotext wrapper)
│   ├── middleware/       # HTTP middleware (auth, logging, etc.)
│   ├── models/           # GORM database models
│   ├── repositories/     # Database access helpers and query-oriented data operations
│   ├── routes/           # Route definitions
│   ├── server/           # HTTP server setup and render helpers
│   └── utils/            # Shared utilities (mailer, DB connection, quota, vacation helpers)
├── locales/              # GNU Gettext .po translation files
├── web/
│   ├── static/           # Static assets (CSS, JS, images)
│   └── templates/        # HTML templates (Go html/template)
├── config.toml.example   # Example configuration file
└── Makefile              # Build and development commands
```

---

## ⚙️ Configuration

Copy `config.toml.example` to `config.toml` and adjust as needed. Key sections:

```toml
[database]
url    = "user:password@tcp(host:3306)/dbname?parseTime=True"
# driver defaults to "mysql"; use "postgres" for PostgreSQL
debug  = false

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

For build commands, CLI usage, and local workflow details, see [DEVELOPMENT.md](DEVELOPMENT.md).

---

## ✅ Change Checklist

Before opening a pull request, verify the items relevant to your change:

- backend validation and frontend UX still match
- new flash messages are written in English in Go code
- new user-facing strings are added to the `.po` files
- templates, handlers, and JS stay aligned for any form flow you changed
- handlers, repositories, and models stay aligned when changing data access logic
- setup or config changes are documented in the correct markdown files

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

## 🔐 Password Policy

Backend password validation is enforced centrally in the handlers layer, while password UI behavior lives in shared frontend code. Passwords must contain:

- at least 8 characters
- at least one uppercase letter
- at least one lowercase letter
- at least one number
- at least one special character

If you change this policy, update both the backend validation and the frontend password generator/UX.

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

For the current build, Makefile, and CLI command reference, see [DEVELOPMENT.md](DEVELOPMENT.md).

---

## 📜 Code of Conduct

Please be respectful and professional in all interactions within the project.

---

Happy coding! 🚀
