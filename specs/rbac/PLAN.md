# Plan — rbac

## Visão Geral

Implementação em 4 fases independentes e incrementais. Cada fase pode ser mergeada separadamente sem quebrar o sistema em produção. A feature flag `rbac.enabled` garante rollout seguro.

---

## Fase 1 — Schema e Modelos (DB + Go)

**Objetivo**: Criar as tabelas RBAC e os modelos GORM sem alterar nenhum comportamento existente.

### Tarefas

1. Criar `internal/models/rbac.go` com `RBACRole`, `RBACPermission`, `RBACAdminRole`
2. Criar migration SQL em `internal/database/migrations/rbac.sql`
3. Registrar as novas structs no auto-migrate do banco (ou migration manual)
4. Criar `internal/rbac/seed.go` — insere roles system e catálogo de permissions via `INSERT IGNORE`
5. Chamar o seed no startup (após `db.AutoMigrate` ou migration manual)
6. Adicionar subcomando `migrate rbac` na CLI existente

**Critério de aceite**: `go run main.go migrate rbac` cria as 4 tabelas e popula seed sem erro. Nenhum comportamento de auth alterado.

---

## Fase 2 — Claims JWT Evoluídos

**Objetivo**: Incluir `permissions` e `roles` no access token sem quebrar tokens existentes.

### Tarefas

1. Estender `auth.Claims` com `Permissions []string` e `Roles []string`
2. Criar `internal/rbac/resolver.go` — dado um `username`, consulta `rbac_admin_roles` → roles → permissions e retorna `[]string`
3. Atualizar `GenerateAccessToken` para receber e incluir permissions/roles
4. Atualizar `handlers/auth_handlers.go` (`AuthLogin`, `AuthRefresh`) para chamar o resolver antes de gerar o token
5. Garantir que `superadmin=true` resulta em `permissions: ["*"]` (bypass simbólico)
6. Manter backward-compat: admins sem roles atribuídos recebem permissions derivadas do `superadmin` flag e `domain_admins` (comportamento atual)

**Critério de aceite**: JWT de superadmin contém `permissions: ["*"]`. JWT de domain_admin contém as permissions do role `domain_admin`. Token ainda válido em todos os handlers existentes.

---

## Fase 3 — Middleware e Aplicação nas Rotas

**Objetivo**: Enforçar as permissions nas rotas sem nenhum handler saber de RBAC.

### Tarefas

1. Criar `internal/middleware/rbac.go` com `RequirePermission` e `RequireRole`
2. Implementar lógica: `superadmin=true` → pass-through; senão checar `Claims.Permissions`
3. Implementar domain scoping: verificar se o recurso alvo está nos `Claims.Domains`
4. Adicionar `RequirePermission` a cada grupo de rotas em `internal/routes/routes.go`
5. Feature flag `rbac.enabled` (viper): quando false, `RequirePermission` é no-op
6. Testes unitários para `RequirePermission` (superadmin bypass, permission match, domain scope, missing)

**Mapeamento de permissions por rota:**

| Rota                  | Permission exigida        |
|-----------------------|---------------------------|
| `GET /domains`        | `domains:read`            |
| `POST /domains`       | `domains:write`           |
| `DELETE /domains/:d`  | `domains:delete`          |
| `GET /mailboxes`      | `mailboxes:read`          |
| `POST /mailboxes`     | `mailboxes:write`         |
| `DELETE /mailboxes/*` | `mailboxes:delete`        |
| `GET /aliases`        | `aliases:read`            |
| `POST /aliases`       | `aliases:write`           |
| `DELETE /aliases/*`   | `aliases:delete`          |
| `GET /alias-domains`  | `alias_domains:read`      |
| `POST /alias-domains` | `alias_domains:write`     |
| `DELETE /alias-d/*`   | `alias_domains:delete`    |
| `GET /admins`         | `admins:read`             |
| `POST /admins`        | `admins:write`            |
| `DELETE /admins/*`    | `admins:delete`           |
| `GET /transports`     | `transports:read`         |
| `POST /transports`    | `transports:write`        |
| `DELETE /transports/*`| `transports:delete`       |
| `GET /dashboard`      | `dashboard:read`          |
| `GET /logs`           | `logs:read`               |
| `GET /maillog`        | `logs:read`               |
| `*/settings/apikeys`  | `settings:read/write`     |

**Critério de aceite**: Admin com role `viewer` consegue listar recursos mas recebe 403 em mutações. Superadmin passa em todos. Feature flag desabilitada = comportamento atual inalterado.

---

## Fase 4 — API de Gerenciamento e Frontend

**Objetivo**: Expor CRUD de roles e atribuições via REST + UI no frontend Vue 3.

### Tarefas — Backend

1. Criar `internal/repositories/rbac.go` — queries GORM para roles/permissions/admin_roles
2. Criar `internal/handlers/rbac_handlers.go` — handlers REST com anotações Swagger
3. Criar `internal/api/dto/rbac.go` — DTOs de request/response
4. Registrar rotas `/api/v1/rbac/*` protegidas por `superadmin` ou `settings:write`
5. Log de auditoria: toda mutação de role chama `LogAction`
6. Adicionar subcomando CLI `rbac assign <username> <role> [domain]`

### Tarefas — Frontend

1. Página `frontend/src/views/RoleManagement.vue` — lista e cria roles customizados
2. Componente `RoleAssignment.vue` — integrado ao formulário de edição de admin
3. Guards de rota em `router/index.ts` baseados em `store/auth.ts` (permissions do JWT)
4. Atualizar `store/auth.ts` para expor `hasPermission(perm: string): boolean`

**Critério de aceite**: Via UI, superadmin cria role `log_viewer`, atribui a um admin, esse admin consegue acessar `/dashboard` e `/logs` mas não criar domínios.

---

## Ordem de Dependência

```
Fase 1 (Schema)
    ↓
Fase 2 (Claims JWT)
    ↓
Fase 3 (Middleware + Rotas)    ← pode ser paralelo com início da Fase 4 backend
    ↓
Fase 4 (API + Frontend)
```

---

## Riscos e Mitigações

| Risco | Mitigação |
|-------|-----------|
| Tokens existentes sem `permissions` no claim | ValidateToken aceita claims sem o campo (zero value = `[]`); middleware recai em comportamento legado quando `rbac.enabled=false` |
| Seed falha em banco com charset diferente | `INSERT IGNORE` é idempotente; seed é re-executado no startup sem efeito colateral |
| Domain scoping inconsistente com `domain_admins` | Fase 2 deriva permissions do `domain_admins` para admins sem roles explícitos |
| Performance: resolver lê DB no login | Resolver usa query com JOIN (uma única query), resultado embedado no JWT — zero overhead por request |
