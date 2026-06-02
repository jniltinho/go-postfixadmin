# Done — rbac

## Fase 1 — Schema e Modelos ✅

- [x] `internal/models/rbac.go` — structs `RBACRole`, `RBACPermission`, `RBACAdminRole` com GORM tags e índice único composto em `rbac_admin_roles`
- [x] `internal/database/database.go` — função `MigrateRBAC()` separada de `MigrateDB()`
- [x] `internal/rbac/seed.go` — seed idempotente de 6 roles system + 24 permissions (inclui `profile:read/write`) via `clause.OnConflict`
- [x] `cmd/migrate_import.go` — subcomando `migrate rbac` (cria tabelas + seed)
- [x] `config.toml`, `config.toml.example`, `web/files/config.default.toml` — seção `[rbac]` com `enabled = false`

## Fase 2 — Claims JWT Evoluídos ✅

- [x] `internal/auth/jwt.go` — campos `Permissions []string` e `Roles []string` em `Claims`; struct `TokenParams` substitui 4+ parâmetros em `GenerateAccessToken`/`GenerateRefreshToken`
- [x] `internal/rbac/resolver.go` — `ResolvePermissions()`, `HasPermission()`, `HasRole()`; superadmin retorna `["*"]`; fallback backward-compat: admins com `domain_admins` mas sem roles RBAC recebem permissions do role `domain_admin` automaticamente
- [x] `internal/handlers/auth_handlers.go` — `AuthLogin` e `AuthRefresh` chamam o resolver; fallback silencioso se tabelas RBAC não existirem
- [x] `internal/middleware/jwt.go` — path de API key também resolve permissions via `resolveAPIKeyClaims()`

## Fase 3 — Middleware e Rotas ✅

- [x] `internal/middleware/rbac.go` — `RequirePermission(perms ...string)` e `RequireRole(roles ...string)` com feature flag `rbac.enabled` e superadmin bypass
- [x] `internal/routes/routes.go` — `RequirePermission` aplicado em todos os 30+ endpoints; `GET/PUT /admins/:username` aceitam `admins:read OR profile:read` / `admins:write OR profile:write`

## Fase 3.5 — Auto-proteção ✅

- [x] `UpdateAdminV1` bloqueia `active=false` e `superadmin=false` quando `targetUsername == claims.Username`
- [x] `RemoveAdminRole` bloqueia remoção do role `superadmin` de si mesmo
- [x] Roles com `system=true` não podem ser deletados ou ter permissions alteradas via API

## Fase 4 — API REST ✅

- [x] `internal/api/dto/rbac.go` — 7 DTOs tipados
- [x] `internal/repositories/rbac.go` — 9 funções de CRUD (`ListRoles`, `GetRoleByID`, `CreateRole`, `UpdateRole`, `DeleteRole`, `ListPermissions`, `ListAdminRoles`, `AssignRole`, `RemoveAdminRole`)
- [x] `internal/handlers/rbac_handlers.go` — 9 handlers REST com anotações Swagger completas e `LogAction` em toda mutação

## Correções pós-deploy ✅

- [x] Permissões `profile:read` e `profile:write` adicionadas ao seed e inseridas no banco — permite que domain_admin veja e edite a própria conta
- [x] Resolver com fallback backward-compat — admins com `domain_admins` mas sem `rbac_admin_roles` recebem permissions automaticamente (sem re-login forçado)
- [x] Role `domain_admin` inserido manualmente no banco para admins existentes via SQL

## Frontend Vue 3 ✅

- [x] `store/auth.ts` — decode JWT no login/refresh; campos `permissions` e `roles` no estado; getters `hasPermission(perm)` e `hasRole(...roles)`; backfill de tokens antigos via `initFromStorage`
- [x] `router/index.ts` — rota `/roles` com guard `requirePermission: 'settings:write'`; `beforeEach` verifica permissão e redireciona para Dashboard se negado
- [x] `SidebarNav.vue` — item "Roles" com ícone `shield` exibido apenas para quem tem `settings:write`; `Transport List` usa o mesmo mecanismo de `isSettingsDisabled`
- [x] `pages/rbac/RoleManagementPage.vue` — grid de roles com badges de permissions coloridas (read/write/delete); modal de criação com seletor de permissions; exclusão com `ConfirmDialog`
- [x] `components/RoleAssignment.vue` — painel integrado ao `AdminEditModal`; lista assignments atuais; modal de atribuição com seleção de role e escopo de domínio; botão de remoção
- [x] `pages/admins/AdminEditModal.vue` — seção `<RoleAssignment>` inserida acima de "Assigned Domains" (visível apenas para superadmin)

## CLI ✅

- [x] `cmd/rbac.go` — subcomandos `rbac assign <username> <role> [domain]` e `rbac seed-existing` (migra todos os admins com `domain_admins` mas sem RBAC roles)

## Scripts ✅

- [x] `DOCUMENTS/rbac_migrate_existing_admins.sql` — SQL idempotente (INSERT IGNORE) para atribuir `domain_admin` a todos os admins existentes com `domain_admins`

## Pendente

- [ ] Testes unitários para `RequirePermission` (superadmin bypass, wildcard, match, feature flag off)
