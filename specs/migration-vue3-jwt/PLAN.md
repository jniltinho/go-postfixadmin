# Plan — migration-vue3-jwt

## Fases de Migração

### Fase 0 — Foundations ✅
- [x] Add `golang-jwt/jwt/v5` + config (`jwt_secret`, `jwt_access_ttl`)
- [x] Criar `internal/auth/jwt.go` (generate/validate)
- [x] Middleware dual-mode: session (legado) + JWT (novo) durante transição

### Fase 1 — API Surface Limpa ✅
- [x] DTOs + formato padronizado de resposta em `internal/api/dto/`
- [x] Portar handlers `*API` para `/api/v1` com `c.Bind` JSON
- [x] Endpoints de auth: `/auth/login`, `/auth/refresh`, `/auth/logout`, `/auth/me`
- [x] Rate limiting no login

### Fase 2 — Frontend Scaffold + Design System ✅
- [x] Scaffold Vue 3 + Vite em `frontend/`
- [x] Design system neo-brutalist (Tailwind CSS v4 + Lucide)
- [x] Layouts: `MainLayout.vue` (sidebar admin) + login
- [x] Auth store (Pinia) + axios interceptors com refresh automático

### Fase 3 — Feature Parity Pages ✅
- [x] Dashboard
- [x] Domains (CRUD + modais)
- [x] Mailboxes (CRUD + modais)
- [x] Aliases (CRUD + modais)
- [x] Alias Domains (CRUD + modais)
- [x] Admins (CRUD + modais)
- [x] Transports (CRUD + modais)
- [x] Logs + Maillog (paginados)
- [x] User portal: forwarding, senha, vacation
- [x] API Keys

### Fase 4 — Build Integration ✅
- [x] Makefile: target `frontend-build` integrado ao `build`
- [x] Dockerfile multi-stage: Node (frontend) → Go build → Alpine
- [x] `//go:embed web/dist locales`
- [x] `internal/server/spa.go`: SPAFileServer com history fallback

### Fase 5 — Cutover + Cleanup ✅
- [x] Default para SPA (remover/flag rotas legadas HTML)
- [x] Remover jQuery, DataTables, templates antigos
- [ ] DEVELOPMENT.md atualizado com novo workflow
- [ ] README atualizado

---

## PR Plan (resumo)

| PR  | Título                                      | Status |
|-----|---------------------------------------------|--------|
| 01  | feat(auth): JWT config + library skeleton   | ✅     |
| 02  | feat(auth): dual middleware session + JWT   | ✅     |
| 03  | feat(api): auth endpoints + rate limiting   | ✅     |
| 04  | refactor(api): dto package + error format   | ✅     |
| 04b | feat(i18n): sync script + consolidate lang  | ⏳     |
| 05  | feat(api): /domains CRUD                    | ✅     |
| 06  | feat(api): /mailboxes CRUD                  | ✅     |
| 07  | feat(api): /aliases CRUD                    | ✅     |
| 08  | feat(api): /alias-domains CRUD              | ✅     |
| 09  | feat(api): /admins CRUD                     | ✅     |
| 10  | feat(api): /transports CRUD                 | ✅     |
| 11  | feat(api): /logs + /maillog                 | ✅     |
| 12  | feat(api): /user/* self-service             | ✅     |
| 13  | feat(api): /settings/apikeys                | ✅     |
| 14a | feat(frontend): scaffold Vue 3 + Vite       | ✅     |
| 14b | feat(frontend): neo-brutalist design system | ✅     |
| 15  | feat(frontend): login + JWT integration     | ✅     |
| 16  | feat(frontend): MainLayout + Dashboard      | ✅     |
| 17-23 | feat(frontend): páginas CRUD             | ✅     |
| 24  | build: frontend build + embed web/dist      | ✅     |
| 25  | feat(server): SPA static + fallback         | ✅     |
| 26  | chore: cutover para SPA, cleanup legado     | ✅     |
| 27  | docs: README + DEVELOPMENT.md              | ⏳     |
