# Migration Plan Summary — go-postfixadmin (Vue 3 + Quasar SPA + JWT API)

**Document**: grok-design-summary-6452fe9e  
**Date**: 2026-05-28  
**Status**: Draft

## Executive Summary

Convert the current Echo + Go templates + jQuery + session-based UI (heavy neo-brutalist Tailwind design) into:
- **Backend**: Clean REST/JSON API (`/api/v1/*`) protected by short-lived JWT access tokens + httpOnly refresh cookies.
- **Frontend**: Vue 3 SPA using Quasar Framework v2, built assets placed in `web/dist` and **embedded** into the Go binary (preserving single-binary deployment).
- **Visual fidelity**: 100% match to current brand (`#3B82F6`, `#6366ee`, `#F97316`, `#1E293B` thick borders + hard shadows, Fira fonts, Lucide icons, uppercase tracking-widest, dot patterns).

All existing functionality (dual admin + user portals, domain scoping, password policy + bcrypt $2y$, vacation+Sieve, logs/maillog, fetchmail, transports, i18n in 3 languages, CLI, packaging) must be preserved with zero regression.

## Why This Migration

Current stack (`main.go`, `internal/server/{server,render}.go`, `routes/routes.go`, `middleware/auth.go`, 49 handler methods, `web/templates/`, `web/static/css/input.css`, jQuery DataTables) is maintainable today but accumulating debt. Existing `*API` handlers (`domain_handlers.go:60+`, `mailbox_handlers.go:74+`, etc.) provide a foundation but are still session-tied and use form values.

Modern SPA + JWT improves DX, security posture, and enables future clients while keeping the beloved single-binary model.

## Core Architecture Decisions

- **Auth**: JWT with claims (`sub`, `type:"admin"|"mailbox"`, `superadmin`, `domains[]`). Access 15-30min (Bearer) + 7d httpOnly refresh cookie (rotation recommended).
- **Routing**: `/api/v1/*` = JSON API (strict). All other routes = SPA `index.html` fallback (history mode).
- **Embedding**: `//go:embed web/dist locales web/static`. New SPA serving middleware (API first, then static, then SPA fallback).
- **Frontend location**: `web/frontend/` (Quasar) → builds to `web/dist`.
- **Theming**: Quasar brand colors + heavy custom `.neo-*` CSS layer replicating exact 457+ brutalist patterns from current templates.
- **i18n**: Hybrid — keep `.po`/gotext for backend; `vue-i18n` JSON (synced keys) for SPA.
- **Build**: Extend Makefile + Dockerfile with Node/Quasar stage.

## Key Technical Areas (Cited)

- Auth dual-mode transition in `internal/middleware/auth.go`.
- Repositories already handle scoping (`repositories/domain_admins.go:9`, `domains.go:12`).
- Password validation & hashing centralized (`handlers/password_helpers.go`, `utils/password.go`).
- Vacation Sieve sync must remain callable from new API (`utils/vacation.go:156`).
- DataTables server processing (logs/maillog) will be replaced by clean paginated QTable endpoints.

## Incremental Rollout (No Big Bang)

**Recommended order** (small, mergeable PRs):

1. JWT config + library + basic middleware (dual session+JWT support).
2. Standardized DTOs + error format + `/api/v1/auth/*` (login/refresh/me).
3. Port existing `*API` methods to proper v1 JSON under `/api/v1` (one resource per PR: domains, mailboxes, aliases, admins, etc.).
4. User self-service APIs (password/forward/vacation + Sieve).
5. Quasar scaffold + pixel-perfect neo-brutalist theme foundation.
6. SPA login + auth store + layouts (admin sidebar vs user header).
7. Feature pages (parallelizable after core).
8. Build/embedding/serving integration (Makefile + Dockerfile + Go SPA fallback).
9. Cutover (default to SPA, deprecate templates) + docs.

~26 focused PRs total. Legacy UI + sessions can coexist behind flags during transition.

## Major Risks & Mitigations

- **Theme mismatch**: Highest priority. Dedicated early PR + visual diff checklist.
- **Auth edge cases** (refresh loops, revocation on password change): Thorough matrix + proven JWT lib.
- **Scope creep**: Strict "one area per PR" rule.
- **i18n drift**: Sync script + CI guard.
- **Build/packaging breakage**: Early Dockerfile + packaging PRs with full verification.

## Alternatives Considered (and Rejected)

1. HTMX + Alpine (lighter, stays in Go) — does not meet explicit "Vue 3 + Quasar SPA" requirement.
2. Separate React frontend repo — destroys single-binary advantage and packaging model.
3. Status-quo jQuery forever — increases long-term debt.

## Open Questions (to resolve pre-impl)

- Exact refresh token revocation (DB allowlist vs rotation only)?
- Auto-detect login type vs explicit `type` field?
- Quasar output dir convention (`web/dist` direct vs copy step)?
- Server-side vs client-only API error message localization?

## Key Files to Touch (High-Level)

**Backend**: `go.mod`, config, `internal/auth/`, `internal/api/dto/`, updated handlers + routes, new SPA serving code, middleware, Makefile, Dockerfile.

**Frontend**: Entire new `web/frontend/` Quasar tree (theming critical in `quasar.config.js` + `src/css/neo-brutalist.css`).

**Docs**: README, DEVELOPMENT.md, new migration notes.

## Next Immediate Steps

1. Approve this design document.
2. Resolve open questions.
3. Start with PR 01 (JWT foundations).

**Full detailed design**: See `/tmp/grok-design-doc-6452fe9e.md` (contains Mermaid diagrams, concrete endpoint examples, exact code patterns for embedding, exhaustive PR breakdown with file lists, theme CSS strategy, security analysis, etc.).

This summary is intentionally concise. The full document is the authoritative reference for implementation.
