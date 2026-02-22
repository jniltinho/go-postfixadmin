---
name: go-postfixadmin
description: >
  Development skill for Go-Postfixadmin — a professional email administration panel built with
  Go 1.24+, Echo v5, GORM v2, html/template (views/), TailwindCSS (public/), MySQL or PostgreSQL,
  Cobra CLI, and Viper TOML config. Assets embedded via go:embed.
  Use this skill whenever writing, reviewing, or refactoring any Go code for Go-Postfixadmin.
  Triggers on: handlers, models, routes, views, CLI commands, config, middleware, utils, or any .go file.
  Especially when code is repetitive — guides extraction into utils/ to keep handlers thin.
  Always use before generating new files, adding features, fixing handlers, or refactoring.
---

# Go-Postfixadmin — Golang Development Skill

## Stack

| Layer | Technology |
|---|---|
| Language | Go 1.24+ |
| Web Framework | Echo v5 (`github.com/labstack/echo/v5`) |
| ORM | GORM v2 |
| Templating | `html/template` — files in `views/` |
| CSS / Assets | TailwindCSS — built to `public/`, watched via `make watch-css` |
| Asset delivery | `go:embed` — views + public baked into binary |
| Database | MySQL 8+ (primary) or PostgreSQL (driver selectable via `--db-driver`) |
| CLI | Cobra (`github.com/spf13/cobra`) |
| Config | Viper (`github.com/spf13/viper`) + TOML (`config.toml`) |
| Binary name | `postfixadmin` |

---

## Project Structure (real)

```
go-postfixadmin/
├── main.go                        # Entry point — calls Cobra root command
├── go.mod / go.sum
├── config.toml.example            # Config template — copy to config.toml
├── Makefile                       # build-prod, build-docker, run, watch-css, deps, tidy, clean
├── Dockerfile                     # Multi-stage: Node (CSS) → Go build → Alpine final (~14MB)
├── tailwind.config.js
├── package.json
├── .github/workflows/
├── DOCUMENTS/
│   ├── screenshots/
│   └── setup/
│       ├── README.md              # Full mail server setup guide
│       └── postfixadmin.service   # Systemd service (deploys to /opt/go-postfixadmin)
├── admin/                         # Admin CLI utilities logic
├── cmd/                           # Cobra subcommands
│   ├── server                     # "server" — starts Echo (flag: --port)
│   ├── migrate                    # "migrate" — runs DB migration
│   ├── importsql                  # "importsql" — imports SQL file to DB
│   ├── admin                      # "admin" — admin management utilities
│   └── version                    # "version" — display version info
├── internal/
│   ├── handlers/
│   │   ├── handlers.go            # Base handler struct + shared deps
│   │   ├── admin_handlers.go
│   │   ├── alias_handlers.go
│   │   ├── alias_domain_handlers.go
│   │   ├── dashboard_handlers.go
│   │   ├── domain_handlers.go
│   │   ├── fetchmail_handlers.go
│   │   ├── mailbox_handlers.go
│   │   └── user_handlers.go
│   ├── middleware/
│   │   └── auth.go
│   ├── models/
│   │   └── models.go
│   ├── routes/
│   │   └── routes.go
│   ├── server/
│   │   ├── server.go              # Echo setup, middleware, startup
│   │   └── render.go              # Template rendering helpers
│   └── utils/
│       ├── db_setup.go
│       ├── db_import.go
│       ├── domain.go
│       ├── logger.go
│       ├── password.go
│       ├── password_generator.go
│       ├── password_test.go
│       ├── permissions.go
│       └── quota.go
├── public/                        # Static assets (CSS, JS, images) — served at /public/*
└── views/                         # html/template files — served from embedded FS
    └── users/
        ├── add_admin.html
        ├── add_alias.html
        ├── add_alias_domain.html
        ├── add_domain.html
        ├── add_fetchmail.html
        ├── add_mailbox.html
        ├── admins.html
        ├── alias_domains.html
        ├── aliases.html
        ├── dashboard.html
        ├── domains.html
        ├── edit_admin.html
        ├── edit_alias.html
        ├── edit_alias_domain.html
        └── ...
```

---

## CLI Commands & Flags

Binary name: `postfixadmin`

**Global (persistent) flags:**

| Flag | Description |
|---|---|
| `--config string` | Config file path (default: `./config.toml`) |
| `--db-driver string` | Database driver: `mysql` or `postgres` |
| `--db-url string` | Full DB connection string (overrides config.toml) |
| `--generate-config` | Write a default `config.toml` to current directory |

**Subcommands:**

| Command | Description |
|---|---|
| `server` | Start the HTTP server (`--port=8080`) |
| `migrate` | Run database AutoMigrate |
| `importsql` | Import a SQL file into the database |
| `admin` | Admin management utilities (see below) |
| `version` | Print version information |

**Admin subcommand flags:**

| Flag | Description |
|---|---|
| `--list-admins` | List all superadmins |
| `--list-domains` | List all domains |
| `--list-mailboxes` | List all mailboxes |
| `--list-aliases` | List all aliases |
| `--domain-admins` | List domain admins |
| `--add-superadmin` | Create superadmin (`"email:password"` or `"email"` for random pass) |

---

## Config (config.toml)

Based on `config.toml.example`. Viper reads this file and env vars override it.

Key sections: `[server]` (port, debug), `[database]` (driver, url or host/port/user/pass/name), `[app]` (env, name, version).

**MySQL DSN** (used in `db_setup.go`):
```
user:password@tcp(host:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
```
For `importsql`, append `&multiStatements=true`.

**PostgreSQL DSN**:
```
host=localhost user=gorm password=gorm dbname=gorm port=5432 sslmode=disable TimeZone=UTC
```

**Priority order:** `--db-url` flag → env var → `config.toml` → default.

---

## Build & Dev Workflow

| Command | What it does |
|---|---|
| `make deps` | `go mod download` + `npm install` |
| `make build-prod` | Tailwind CSS build + Go binary compile |
| `make build-docker` | Multi-stage Docker image (~14MB final, UPX compressed) |
| `make run` | Build + start server locally |
| `make watch-css` | Tailwind watcher for UI development |
| `make clean` | Remove binary and generated CSS |
| `make tidy` | `go mod tidy` |

Deploy to Linux: copy binary to `/opt/go-postfixadmin/`, place `config.toml` there, use `DOCUMENTS/setup/postfixadmin.service` for systemd.

---

## Echo v5 — Key Differences from v4

- Import: `github.com/labstack/echo/v5`
- `echo.Map` **removed** — always use `map[string]any`
- Path params: `c.PathParam("id")`, not `c.Param("id")`
- Static files: `echo.StaticDirectoryHandler`, not `e.Static()`
- `e.Renderer` interface **removed** — rendering done manually in `server/render.go`
- Middleware: `echo/v5/middleware` (same module)

---

## go:embed Rules

- `ViewsFS` and `PublicFS` declared as `embed.FS` (in `main.go` or a dedicated `embed.go`)
- Always `fs.Sub()` to strip top-level prefix before `ParseFS` or `StaticDirectoryHandler`
- Never `os.DirFS("views")` or `ParseGlob(...)` — filesystem not available at runtime
- View template names after `fs.Sub` are e.g. `"users/dashboard.html"` (no `views/` prefix)
- Public assets served at `/public/*`

---

## ⭐ The Golden Rule: Keep Handlers Thin

Handler files are split by entity (`admin`, `alias`, `alias_domain`, `dashboard`, `domain`, `fetchmail`, `mailbox`, `user`). Each handler function must stay under ~30 lines.

A correct handler: parses input → calls a util or DB function → renders output. That's it.

**If logic appears in 2+ handlers, extract it:**

| Repeated pattern | Where to put it |
|---|---|
| Template rendering | `server/render.go` |
| Permission check | `utils/permissions.go` |
| Password operation | `utils/password.go` |
| Quota logic | `utils/quota.go` |
| Domain validation | `utils/domain.go` |
| GORM query | `utils/db_setup.go` or a new `utils/repo_*.go` |
| JSON/HTML error response | `server/render.go` (add `RenderError`) |

---

## GORM + MySQL Conventions

- Use `gorm.Model` for `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt`
- `TableName()` must match the exact Postfix MySQL schema table names (e.g. `mailbox`, `alias`, `domain`, `fetchmail`)
- MySQL type tags: `size:255` → VARCHAR, `type:text` → TEXT, `type:decimal(10,2)` → quotas, `type:tinyint(1)` → active/enabled booleans
- DB functions return `(nil, nil)` for not-found — never propagate `gorm.ErrRecordNotFound` to handlers
- Connection pool in `db_setup.go`: `SetMaxIdleConns(10)`, `SetMaxOpenConns(100)`, `SetConnMaxLifetime(time.Hour)`

---

## RBAC — Access Control

Two roles: **Superadmin** (manages everything) and **Domain Admin** (manages only their domains/mailboxes/aliases).

- `utils/permissions.go` holds all role checks
- `internal/middleware/auth.go` enforces authentication on protected routes
- `admin/` package holds the CLI admin utilities (create superadmin, list entities, etc.)

---

## Views Conventions

- Files under `views/users/` — template names use this subdirectory: `"users/dashboard.html"`
- Assets at `/public/css/output.css` (generated by Tailwind CLI, embedded)
- Tailwind classes in HTML files only — never in Go code
- Flash messages rendered in base layout via `.Flash.TailwindClass` and `.Flash.Message`

---

## Refactoring Checklist

Before committing any handler:

- [ ] Handler body under 30 lines
- [ ] `c.PathParam("id")` not `c.Param("id")` (Echo v5)
- [ ] `map[string]any` not `echo.Map` (Echo v5)
- [ ] View path includes subdir: `"users/dashboard.html"`
- [ ] Asset paths use `/public/...`
- [ ] Views/public served from embedded FS — not `os.DirFS`
- [ ] Rendering via `server/render.go` — not inline
- [ ] Password ops via `utils/password.go`
- [ ] Permission checks via `utils/permissions.go`
- [ ] Quota logic via `utils/quota.go`
- [ ] No inline GORM queries in handlers
- [ ] Errors rendered via `RenderError` in `server/render.go`
- [ ] MySQL GORM tags: `size:255`, `type:text`, `type:tinyint(1)`, `type:decimal(10,2)`
- [ ] Table names match Postfix schema exactly

> 📖 Utils contracts → `references/utils.md`
