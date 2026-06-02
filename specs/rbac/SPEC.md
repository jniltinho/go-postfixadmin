# Spec — rbac

**Objetivo**: Introduzir controle de acesso baseado em papéis (RBAC) granular, substituindo a lógica binária `superadmin/domain_admin` por roles e permissions persistidos no banco de dados.

**Status**: 🔲 Planejado

---

## Contexto e Motivação

O sistema atual possui apenas dois níveis de acesso:

- `superadmin=true` → acesso total
- `superadmin=false` → acesso restrito aos domínios atribuídos via `domain_admins`

Essa granularidade é insuficiente quando múltiplos admins precisam de escopos distintos (ex: um admin que só lê relatórios, outro que gerencia somente mailboxes). O RBAC resolve isso com roles reutilizáveis e permissions atômicas, mantendo backward-compatibility total com o schema atual do PostfixAdmin.

---

## Modelo de Dados

### Novas tabelas

```sql
-- Papéis disponíveis no sistema
CREATE TABLE rbac_roles (
    id          INT AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(64)  NOT NULL UNIQUE,
    description VARCHAR(255) NOT NULL DEFAULT '',
    system      TINYINT(1)   NOT NULL DEFAULT 0,   -- 1 = built-in, não deletável
    created     DATETIME     NOT NULL DEFAULT '2000-01-01 00:00:00',
    modified    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Permissões atômicas
CREATE TABLE rbac_permissions (
    id       INT AUTO_INCREMENT PRIMARY KEY,
    name     VARCHAR(128) NOT NULL UNIQUE,   -- ex: "mailboxes:write"
    resource VARCHAR(64)  NOT NULL,          -- ex: "mailboxes"
    action   VARCHAR(32)  NOT NULL           -- ex: "write"
);

-- Relacionamento N:N entre roles e permissions
CREATE TABLE rbac_role_permissions (
    role_id       INT NOT NULL,
    permission_id INT NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id)       REFERENCES rbac_roles(id)       ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES rbac_permissions(id) ON DELETE CASCADE
);

-- Atribuição de roles a admins (opcionalmente scoped por domínio)
CREATE TABLE rbac_admin_roles (
    id         INT AUTO_INCREMENT PRIMARY KEY,
    username   VARCHAR(255) NOT NULL,
    role_id    INT          NOT NULL,
    domain     VARCHAR(255) NOT NULL DEFAULT '',  -- '' = global (todos os domínios do admin)
    created    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_admin_role_domain (username, role_id, domain),
    FOREIGN KEY (role_id) REFERENCES rbac_roles(id) ON DELETE CASCADE
);
```

### Roles pré-definidos (system=1, seed automático)

| role            | descrição                                        |
|-----------------|--------------------------------------------------|
| `superadmin`    | Acesso total — espelha `Admin.Superadmin=true`   |
| `domain_admin`  | CRUD completo no escopo de domínios atribuídos   |
| `mailbox_admin` | CRUD de mailboxes e aliases (sem domínios)       |
| `alias_admin`   | CRUD apenas de aliases e alias_domains           |
| `viewer`        | Leitura em todos os recursos do seu escopo       |
| `report_viewer` | Apenas dashboard, logs e maillog                 |

### Permissions catalog

| permission              | resource      | action   |
|-------------------------|---------------|----------|
| `domains:read`          | domains       | read     |
| `domains:write`         | domains       | write    |
| `domains:delete`        | domains       | delete   |
| `mailboxes:read`        | mailboxes     | read     |
| `mailboxes:write`       | mailboxes     | write    |
| `mailboxes:delete`      | mailboxes     | delete   |
| `aliases:read`          | aliases       | read     |
| `aliases:write`         | aliases       | write    |
| `aliases:delete`        | aliases       | delete   |
| `alias_domains:read`    | alias_domains | read     |
| `alias_domains:write`   | alias_domains | write    |
| `alias_domains:delete`  | alias_domains | delete   |
| `admins:read`           | admins        | read     |
| `admins:write`          | admins        | write    |
| `admins:delete`         | admins        | delete   |
| `transports:read`       | transports    | read     |
| `transports:write`      | transports    | write    |
| `transports:delete`     | transports    | delete   |
| `logs:read`             | logs          | read     |
| `dashboard:read`        | dashboard     | read     |
| `settings:read`         | settings      | read     |
| `settings:write`        | settings      | write    |

---

## Modelos Go (GORM)

```go
// internal/models/rbac.go

type RBACRole struct {
    ID          int       `gorm:"column:id;primaryKey;autoIncrement"`
    Name        string    `gorm:"column:name;type:varchar(64);uniqueIndex"`
    Description string    `gorm:"column:description;type:varchar(255);not null"`
    System      bool      `gorm:"column:system;not null;default:false"`
    Created     time.Time `gorm:"column:created;not null"`
    Modified    time.Time `gorm:"column:modified;not null;autoUpdateTime"`
    Permissions []RBACPermission `gorm:"many2many:rbac_role_permissions"`
}

type RBACPermission struct {
    ID       int    `gorm:"column:id;primaryKey;autoIncrement"`
    Name     string `gorm:"column:name;type:varchar(128);uniqueIndex"`
    Resource string `gorm:"column:resource;type:varchar(64);not null"`
    Action   string `gorm:"column:action;type:varchar(32);not null"`
}

type RBACAdminRole struct {
    ID       int    `gorm:"column:id;primaryKey;autoIncrement"`
    Username string `gorm:"column:username;type:varchar(255);not null;index"`
    RoleID   int    `gorm:"column:role_id;not null"`
    Domain   string `gorm:"column:domain;type:varchar(255);not null;default:''"`
    Created  time.Time `gorm:"column:created;not null;autoCreateTime"`
    Role     RBACRole  `gorm:"foreignKey:RoleID"`
}
```

---

## Claims JWT (evolução)

Os claims JWT ganham o campo `permissions` para evitar consulta ao banco em cada request:

```json
{
  "username": "admin@example.com",
  "type":     "admin",
  "superadmin": false,
  "domains":  ["example.com"],
  "permissions": ["mailboxes:read", "mailboxes:write", "aliases:read"],
  "roles":    ["mailbox_admin"]
}
```

`superadmin=true` continua sendo o bypass de todas as permission checks (backward-compat).

---

## Middleware RBAC

```go
// internal/middleware/rbac.go

// RequirePermission retorna middleware que exige ao menos uma das permissões listadas.
func RequirePermission(perms ...string) echo.MiddlewareFunc

// RequireRole retorna middleware que exige ao menos um dos roles listados.
func RequireRole(roles ...string) echo.MiddlewareFunc
```

Aplicação nas rotas (sem quebrar rota pública nem user portal):

```go
// Apenas superadmin ou quem tem admins:write
protected.POST("/admins", h.CreateAdminV1, middleware.RequirePermission("admins:write"))
protected.DELETE("/admins/:username", h.DeleteAdminV1, middleware.RequirePermission("admins:delete"))

// Viewer consegue listar
protected.GET("/domains", h.ListDomainsV1, middleware.RequirePermission("domains:read"))
```

---

## API REST — Gerenciamento de Roles

Base path: `/api/v1/rbac` — protegida por JWT + `superadmin` ou `settings:write`.

| Método | Endpoint                               | Descrição                        |
|--------|----------------------------------------|----------------------------------|
| GET    | `/rbac/roles`                          | Lista todos os roles             |
| POST   | `/rbac/roles`                          | Cria role customizado            |
| GET    | `/rbac/roles/:id`                      | Detalhe do role (com perms)      |
| PUT    | `/rbac/roles/:id`                      | Edita role (não system)          |
| DELETE | `/rbac/roles/:id`                      | Remove role (não system)         |
| GET    | `/rbac/permissions`                    | Lista todas as permissions       |
| GET    | `/rbac/admins/:username/roles`         | Roles atribuídos a um admin      |
| POST   | `/rbac/admins/:username/roles`         | Atribui role a um admin          |
| DELETE | `/rbac/admins/:username/roles/:roleId` | Remove role de um admin          |

---

## Regras de Auto-proteção

Proteções que impedem um admin de se bloquear acidentalmente. Aplicadas no handler, não no middleware.

### Já implementado (manter)

| Operação                  | Handler atual               | Comportamento                                              |
|---------------------------|-----------------------------|------------------------------------------------------------|
| `DELETE /admins/:username`| `DeleteAdminV1` (linha 517) | 403 se `targetUsername == claims.Username`                 |
| `DELETE /admins/:username`| `DeleteAdmin` (linha 130)   | 403 se `username == loggedInUser` (handler legado SSR)     |

### Lacuna existente (corrigir na Fase 4)

`UpdateAdminV1` ([admin_handlers.go:462](../../internal/handlers/admin_handlers.go#L462)) **não impede** que um superadmin defina `superadmin=false` em si mesmo, o que efetivamente revoga seu próprio acesso total. Adicionar guard:

```go
// Em UpdateAdminV1, antes de aplicar updates:
if req.Superadmin != nil && !*req.Superadmin && targetUsername == claims.Username {
    return dto.Forbidden(c, "you cannot revoke your own superadmin flag")
}
```

### Regras RBAC (novas)

| Operação                                      | Regra                                                                 |
|-----------------------------------------------|-----------------------------------------------------------------------|
| `DELETE /rbac/admins/:username/roles/:roleId` | 403 se `username == claims.Username` e role é `superadmin`           |
| `POST /rbac/admins/:username/roles`           | Permitido auto-atribuição apenas de roles não-superadmin              |
| Qualquer mutação de role `system=true`        | 403 sempre — roles system são imutáveis via API                       |
| Desativar própria conta (`active=false`)      | 403 — guard a adicionar em `UpdateAdminV1` junto com o de superadmin  |

---

## Estratégia de Migração

1. **Sem quebra de schema existente** — apenas novas tabelas adicionadas via `migrate` CLI.
2. **Seed automático** no startup: roles e permissions system são inseridos com `INSERT IGNORE`.
3. **Backward-compat**: admins existentes continuam funcionando. Ao primeiro login após RBAC, o sistema resolve as permissions via `superadmin` flag e `domain_admins` e embute no JWT.
4. **Feature flag** `rbac.enabled` (default `false` na v1) — quando `false`, middleware RBAC é no-op, preservando comportamento atual.
5. **Rollout gradual**: habilitar feature flag por instância via config.

---

## Segurança

- Permissions checadas no middleware (antes do handler) — nunca no handler.
- `superadmin=true` faz bypass de todas as checks (imutável para retrocompat).
- Roles `system=true` não podem ser deletados ou ter suas permissions alteradas via API.
- Mutações de roles logadas via `LogAction` existente.
- Claims JWT expiram; permissions ficam em cache apenas pelo TTL do access token (15 min).
- Domain scoping: permission `mailboxes:write` scoped para `domain=example.com` só vale naquele domínio — middleware cruza com `Domains` do claim.
