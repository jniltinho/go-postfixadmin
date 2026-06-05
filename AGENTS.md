# AGENTS.md

> Entry point for AI coding agents working on **Go-PostfixAdmin**.
> A self-contained Go binary that ships a Vue 3 SPA and REST API for managing Postfix + Dovecot mail servers.

This file is intentionally short. For deep context, follow the **Read these first** links below.

---

## Read these first

| Doc | Purpose |
| --- | --- |
| [README.md](README.md) | What the project is, screenshots, feature list, quick start |
| [DEVELOPMENT.md](DEVELOPMENT.md) | Build commands, CLI reference, frontend dev workflow |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Workflow, conventions, change checklist |
| [FEATURES.md](FEATURES.md) | Capability overview |
| [`.agent/ARCHITECTURE.md`](.agent/ARCHITECTURE.md) | Existing agent kit — specialists, skills, workflows |
| [`.agent/rules/build.md`](.agent/rules/build.md) | Always-on build/run commands for this project |
| [`DOCUMENTS/setup/README.md`](DOCUMENTS/setup/README.md) | Production setup (Postfix, Dovecot, MariaDB/PostgreSQL) |

The repo already ships a 16-agent specialist kit in `.agent/agents/` plus 25+ skills in `.agent/skills/`. **Reuse it** — do not redefine the same roles in this file.

---

## Project snapshot

- **Module**: `go-postfixadmin` (Go 1.26)
- **Backend**: Echo v5, GORM (MySQL + PostgreSQL), Cobra CLI, Viper config, `log/slog`, `swaggo/swag`
- **Auth**: JWT (short-lived access + httpOnly refresh) with RBAC roles/permissions, optional API keys
- **Frontend**: Vue 3 + TS + Vite + Pinia + Tailwind v4 + Lucide icons — **embedded** at build time via `//go:embed`
- **i18n**: GNU Gettext `.po` files loaded with `leonelquinteros/gotext` (en, pt_BR, es)
- **Packaging**: single static binary (UPX-compressed for releases) + multi-stage Alpine Docker image + `.deb` / `.rpm`
- **API surface**: versioned REST at `/api/v1` with Swagger UI at `/swagger/`
- **Entry point**: `main.go` → `cmd.Execute(embeddedFiles)` (Cobra)

---

## Quick commands

> Full reference in [DEVELOPMENT.md](DEVELOPMENT.md). The rule at `.agent/rules/build.md` is **always-on** for build/run tasks.

```bash
# Install deps (Go modules + frontend once)
make deps
cd frontend && npm install && cd ..

# Local dev loop (SPA hot-reload + Go server)
make frontend          # rebuild Vue 3 → web/dist
make build             # full build (frontend + Go binary with embed)
make run               # build + start ./bin/postfixadmin server

# Server only (when frontend is already built and you don't want to re-embed)
./bin/postfixadmin server --port=8080
./bin/postfixadmin server --debug

# Database
./bin/postfixadmin migrate           # core schema
./bin/postfixadmin migrate rbac      # RBAC tables + seed system roles
./bin/postfixadmin importsql file.sql

# Regenerate Swagger after editing // @... annotations
make swagger && make build

# Production artifacts
make build-prod       # static binary + UPX --best --lzma
make build-docker
make deb / make rpm

# Test
go test ./...
```

Generated config: `./bin/postfixadmin --generate-config` → `config.toml` (also see `config.toml.example` and `web/files/config.default.toml`, which is embedded).

---

## Code layout

```
.
├── main.go                      # //go:embed all:web/dist web/files all:locales → cmd.Execute
├── cmd/                         # Cobra commands (root + one file per subcommand)
│   ├── root.go                  # Global flags, Viper init, slog setup
│   ├── server.go                # `server` subcommand — starts Echo HTTP
│   ├── admin.go, mailbox.go, domain.go, transport.go, rbac.go, ...
│   ├── admin/                   # CLI subpackage: cli.go, cli_add.go, ...
│   ├── mailbox/                 # CLI subpackage: cli_add_user.go, cli_import_csv.go, ...
│   └── transport/               # CLI subpackage incl. tcpserver.go (Postfix transport_maps)
├── internal/
│   ├── api/dto/                 # APIResponse, APIError, request/response DTOs, error helpers
│   ├── auth/                    # JWT signing/validation (admin + mailbox token types)
│   ├── database/                # ConnectDB(), MigrateDB(), MigrateRBAC()
│   ├── handlers/                # Echo HTTP handlers — *Handler methods on a single struct
│   ├── middleware/              # auth, jwt, rbac, csrf, ratelimit
│   ├── models/                  # GORM models (sql2go-generated, plus rbac.go)
│   ├── rbac/                    # resolver.go, seed.go (system roles + permission catalog)
│   ├── repositories/            # Data access — one file per aggregate
│   ├── routes/                  # RegisterRoutes() — Echo route table + RBAC guard chain
│   ├── server/                  # StartServer(), template/render, SPA handler
│   ├── utils/                   # mailer, password, maillog_reader, vacation, quota
│   └── i18n/                    # gotext wrapper, embedded .po loader
├── pkg/crypt/                   # Publicly importable helpers
├── frontend/                    # Vue 3 + TS + Vite source (npm run build → web/dist)
├── web/
│   ├── dist/                    # Built SPA — embedded at runtime via //go:embed
│   ├── files/                   # config.default.toml (embedded)
│   ├── static/                  # Legacy assets (do not add new ones here)
│   └── templates/               # Legacy Go html/template UI (still served for some routes)
├── locales/                     # en/, es/, pt_BR/ — each contains default.po
├── docs/                        # Generated by `make swagger` (embedded in binary)
├── DOCUMENTS/                   # Setup guides, screenshots, systemd units
├── Dockerfile, docker-compose.yml, Makefile
├── config.toml.example
└── .agent/                      # Existing AI agent kit — see ARCHITECTURE.md
```

---

## Conventions

### Go style

- **Go 1.26**, no `// TODO: modernize` — use `slices`, `maps`, `cmp`, generic helpers freely.
- One package per concern. Handlers are methods on a single `*Handler{DB *gorm.DB}` in `internal/handlers`.
- Cobra: one file per subcommand in `cmd/`. Subcommand implementation in `cmd/<group>/...` package, wired in `cmd/<group>.go`.
- Use `log/slog` (not zerolog/logrus) for new code. `zerolog` is used **only** in `cmd/transport` and `cmd/transport/tcpserver.go` — keep it there.
- Use `c.JSON(...)` or `c.Render(...)` on `*echo.Context` (Echo v5). For v1 API, always wrap responses in `dto.APIResponse` via `dto.WriteSuccess` / `dto.WriteError`.
- Use `viper.Get*` for config reads. Flag → Viper precedence is already wired in `cmd/root.go` (`viper.BindPFlag`).
- Errors: return them, log with `slog.Error("...", "error", err)`, and `os.Exit(1)` in Cobra Run.
- **No comments** on the code itself unless the surrounding file already uses them to explain non-obvious behavior (e.g. middleware, auth flows). Do not add banner comments to new files.

### Naming

- Files: `snake_case.go`. Test files: `*_test.go` in the same package.
- Exported identifiers use Go stdlib conventions — `*Handler`, `ConnectDB`, `StartServer`, `ParseToken`.
- Cobra commands: short imperative (`server`, `migrate`, `importsql`, `rbac`).
- Handler methods: `ListX`, `GetX`, `CreateX`, `UpdateX`, `DeleteX` for `/api/v1` (V1 suffix) and matching names for legacy HTML routes.

### Frontend

- **Build the SPA from `frontend/`. Never edit `web/dist/`** — it is generated and re-embedded.
- Tailwind v4 utilities, Lucide icons, Pinia stores under `frontend/src/stores`, vue-router under `frontend/src/router`.
- Design aesthetic: **neo-brutalism** (2px/4px borders, high contrast, sharp `neo-shadow`).

### Database / migrations

- New GORM model → add it to `database.MigrateDB()` (or `MigrateRBAC()` for the `rbac_*` tables).
- Models in `internal/models/models.go` are `// Code generated by sql2go. DO NOT EDIT.` — do not hand-edit. Add new models in a new file in the same package.
- DB driver selection: `database.ConnectDB(dsn, driver)` reads `db-driver` / `db-url` flags → `database.driver` / `database.url` in `config.toml` → `DB_DRIVER` / `DB_URL` env vars.

### Auth & RBAC

- Public routes go in `apiV1.Group("", middleware.CSRFOriginCheck(), loginLimiter.Middleware())`.
- All protected routes use `protected := apiV1.Group("", middleware.JWTAuthMiddleware(h.DB))`.
- For new endpoints requiring permissions, append `middleware.RequirePermission(rbac.PermXxxYyy)` — see `rbac.PermDomainsRead` etc. in `internal/rbac/`.
- User portal routes (mailbox self-service) live under `/api/v1/user/*` and do **not** require RBAC permissions.
- Token types: `"admin"` and `"mailbox"`. Embed `permissions` and `roles` in claims so middleware needs no DB round-trip.

### i18n

- New user-facing English string in Go code → add the `msgid` to **all three** `.po` files (`locales/en/default.po`, `locales/es/default.po`, `locales/pt_BR/default.po`).
- Use `T(c, "MessageId")` and `TData(c, "MessageId", map[string]any{...})` helpers in handlers.
- Adding a new language: see [CONTRIBUTING.md § Localization](CONTRIBUTING.md#-localization-i18n).

### API documentation

- Every v1 handler must carry `// @Summary ...` ... `// @Router ...` annotations.
- After editing annotations: `make swagger && make build` (the generated `docs/` is embedded at build time).

---

## Change checklist (mirrors CONTRIBUTING.md)

Before opening a PR, verify:

- [ ] `go test ./...` passes
- [ ] `make build` succeeds (regenerates embedded assets)
- [ ] `make swagger` was re-run if any `// @...` annotation changed
- [ ] New user-facing strings added to all three `.po` files
- [ ] New GORM model registered in `database.MigrateDB()` / `MigrateRBAC()`
- [ ] New permissions registered in `internal/rbac/` and seeded
- [ ] Handler, repository, and model stay aligned
- [ ] README.md / DEVELOPMENT.md / FEATURES.md / setup docs updated when behavior, config keys, CLI flags, or setup steps change
- [ ] Do **not** commit `config.toml`, `web/dist/`, `bin/`, or `frontend/node_modules/`

---

## Common pitfalls

- **Editing generated files.** `web/dist/`, `docs/`, `internal/models/models.go` are generated. Edit the source and regenerate.
- **Forgetting to embed.** Anything new under `web/dist`, `web/files`, or `locales/` is picked up by `//go:embed` in `main.go` — you only need to `make build` again.
- **Using `c.JSON` for errors in v1 endpoints.** Use `dto.WriteError(c, dto.ErrCodeXxx, "message")` so the envelope stays consistent.
- **Skipping RBAC permission guards** on new v1 admin routes. Every admin route must be guarded by `middleware.RequirePermission(...)`; see `internal/routes/routes.go` for the convention.
- **Mixing `log/slog` and `zerolog`.** Default to `slog`. The transport TCP server is the only place that intentionally uses `zerolog` (it sets up its own colored console writer).
- **Hard-coding config values.** Read from Viper (`viper.GetString(...)`, `viper.GetInt(...)`, etc.) so config files and `--flag` overrides work.
- **Hand-editing `internal/models/models.go`.** It is `// Code generated by sql2go. DO NOT EDIT.`

---

## Where to put new code

| You are adding... | Put it in |
| --- | --- |
| New Cobra subcommand | `cmd/<name>.go` (flags, Run) + `cmd/<name>/cli_*.go` (logic) |
| New HTTP handler (v1 JSON) | `internal/handlers/<resource>_handlers.go`, method on `*Handler` |
| Legacy HTML route | `internal/handlers/<resource>_handlers.go` (render with `c.Render`) |
| New Echo route | `internal/routes/routes.go` (with RBAC guard if admin-only) |
| New GORM model | `internal/models/<name>.go` (new file), then add to `MigrateDB()` |
| New repository function | `internal/repositories/<resource>.go` |
| New RBAC permission | `internal/rbac/` (constant + seed entry) |
| New JWT helper | `internal/auth/` |
| New i18n string | `T(c, "Key")` in handler + add to all three `.po` files |
| New middleware | `internal/middleware/<name>.go` (Echo v5 signature: `func(next echo.HandlerFunc) echo.HandlerFunc`) |
| New CLI utility | `internal/utils/` |
| New Vue page / store | `frontend/src/pages/`, `frontend/src/stores/` |
| New embed asset | drop under `web/dist/`, `web/files/`, or `locales/<lang>/` then `make build` |

---

## Don't

- Don't add new legacy templates (`web/templates/`, `web/static/`) — Vue SPA is the primary UI.
- Don't introduce a new logger (no `zap`, `logrus`).
- Don't bypass `RequirePermission` on admin routes, even for "internal" endpoints.
- Don't commit secrets, generated binaries, `web/dist/`, `node_modules/`, or local `config.toml`.
- Don't run `go mod tidy` without re-running `go mod download` to confirm reproducibility.
- Don't disable the rate limiter or CSRF check on auth endpoints.
