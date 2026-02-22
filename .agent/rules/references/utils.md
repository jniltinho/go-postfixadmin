# Utils / Helpers — Reference Guide

Domain-specific helpers live in `internal/utils/`. Rendering helpers live in `internal/server/render.go`.

> **Render functions** (`Render`, `RenderPartial`, `RenderStatus`, `RenderError`, `TemplateFuncs`) live in `internal/server/render.go`. Everything else is in `internal/utils/`.

---

## server/render.go

**`Render(c, name, data)`** — Executes a named template (e.g. `"users/dashboard.html"`) using the `*template.Template` stored in context under key `"tmpl"`. Auto-injects flash via `GetFlash(c)` as `"Flash"`. Sets Content-Type to `text/html; charset=utf-8`, status 200.

**`RenderPartial(c, name, data)`** — Same but skips flash injection. For HTMX partial responses.

**`RenderStatus(c, status, name, data)`** — Same as `Render` with a custom HTTP status. Used internally by `RenderError`.

**`RenderError(c, err)`** — Inspects error type: handles `*AppError` (uses its Code), `gorm.ErrRecordNotFound` (renders 404), all others (logs + renders 500). Always renders `"users/error.html"` or equivalent.

**`TemplateFuncs()`** — Returns `template.FuncMap` registered before `ParseFS`. Includes: `add`, `sub`, `mul` (int math), `formatDate` (Jan 02 2006), `formatDateTime`, `safeHTML` → `template.HTML`, `safeURL` → `template.URL`, `seq(n)` → `[]int{1..n}`.

**`AppError`** — Struct with `Code int`, `Message string`, `Err error`. Implements `error` + `Unwrap`. Constructors: `NewNotFoundError(msg)` (404), `NewBadRequestError(msg)` (400), `NewForbiddenError(msg)` (403), `NewUnauthorizedError(msg)` (401).

---

## utils/db_setup.go

**`Connect(driver, dsn string)`** — Opens GORM connection with the given driver (`mysql` or `postgres`). Configures pool: `SetMaxIdleConns(10)`, `SetMaxOpenConns(100)`, `SetConnMaxLifetime(time.Hour)`. Returns `*gorm.DB`. Panics on failure.

**`Migrate(db *gorm.DB)`** — Runs `db.AutoMigrate` for all models (`Admin`, `Domain`, `Mailbox`, `Alias`, `AliasDomain`, `Fetchmail`). Called once after `Connect`.

---

## utils/db_import.go

**`ImportSQL(db *gorm.DB, filepath string)`** — Reads a SQL file and executes it against the database. Requires DSN with `multiStatements=true`. Used by the `importsql` CLI subcommand.

---

## utils/password.go

**`HashPassword(password string)`** — Bcrypt-hashes a plaintext password. Returns `(hash string, err error)`.

**`CheckPassword(hash, password string)`** — Verifies plaintext against bcrypt hash. Returns `bool`.

**`DovecotHash(password string)`** — Generates a Dovecot-compatible password hash (scheme used by Postfix/Dovecot in the `mailbox` table). Use this when creating or updating mailbox passwords — NOT plain bcrypt.

---

## utils/password_generator.go

**`GeneratePassword(length int)`** — Returns a cryptographically secure random password of the given length using letters, digits, and symbols. Used by `--add-superadmin "email"` (no password given) and mailbox creation with blank password.

---

## utils/permissions.go

**`IsSuperadmin(c echo.Context)`** — Returns `true` if the session user has superadmin role.

**`IsDomainAdmin(c echo.Context, domain string)`** — Returns `true` if the session user is admin for the given domain.

**`RequireSuperadmin(c echo.Context)`** — Returns `*AppError` (403) if user is not superadmin. Call at top of superadmin-only handlers.

**`RequireDomainAdmin(c echo.Context, domain string)`** — Returns `*AppError` (403) if user cannot manage the given domain.

---

## utils/quota.go

**`FormatQuota(mb int64)`** — Converts a quota value in MB to a human-readable string (e.g. `"500 MB"`, `"2 GB"`). 0 means unlimited.

**`ParseQuota(s string)`** — Parses a human-readable quota string into MB as `int64`. Returns error on invalid input.

**`QuotaExceeded(usedMB, limitMB int64)`** — Returns `true` if `usedMB >= limitMB` and `limitMB > 0`. Use before allowing mailbox creation or size updates.

---

## utils/domain.go

**`NormalizeDomain(domain string)`** — Lowercases and trims whitespace. Always normalize before DB writes.

**`IsValidDomain(domain string)`** — Returns `true` if the string is a syntactically valid domain name (no scheme, no path, no port).

**`ExtractDomain(email string)`** — Returns the domain part of an email address. Returns empty string if malformed.

---

## utils/logger.go

**`InitLogger(debug bool)`** — Initializes the global `slog` logger. Debug mode uses `slog.LevelDebug`; production uses `slog.LevelInfo`. Writes structured JSON logs. Call once in `server.Run` or `cmd/server`.

---

## internal/middleware/auth.go

**`AuthMiddleware()`** — Echo middleware that validates the session and rejects unauthenticated requests (redirect to login or 401 JSON). Stores user info in context for downstream handlers and permission utils.

---

## Pagination (if needed)

If pagination is added to any list handler, use a `Pagination` struct with `Page`, `Limit`, `Offset()`, `TotalPages(total int64)`, `HasPrev()`, `HasNext(total)`. Parse from `?page=` and `?limit=` query params with safe defaults (page=1, limit=25, max=100). Extract to `utils/pagination.go`.

---

## Postfix MySQL Schema — Key Table Names

These must match exactly in `TableName()` overrides:

| Model | Table name |
|---|---|
| Domain | `domain` |
| Mailbox | `mailbox` |
| Alias | `alias` |
| AliasDomain | `alias_domain` |
| Admin | `admin` |
| Fetchmail | `fetchmail` |
| DomainAdmins | `domain_admins` |

Postfix schema uses `active TINYINT(1)` for enabled/disabled on most entities. Always include it in models with `gorm:"type:tinyint(1);default:1"`.
