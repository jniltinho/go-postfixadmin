# Features — rbac

## Implementado

_(nenhum item implementado ainda)_

## Pendente / Planejado

### Banco de Dados
- [ ] Tabela `rbac_roles` — definição de papéis
- [ ] Tabela `rbac_permissions` — catálogo de permissões atômicas
- [ ] Tabela `rbac_role_permissions` — relacionamento N:N role ↔ permission
- [ ] Tabela `rbac_admin_roles` — atribuição de roles a admins (com escopo de domínio opcional)
- [ ] Seed automático de roles system (`superadmin`, `domain_admin`, `mailbox_admin`, `alias_admin`, `viewer`, `report_viewer`)
- [ ] Seed automático do catálogo de 22 permissions

### Modelos Go
- [ ] `RBACRole` — struct GORM com associação `many2many` para permissions
- [ ] `RBACPermission` — struct GORM
- [ ] `RBACAdminRole` — struct GORM com FK para `rbac_roles`

### JWT / Auth
- [ ] Adicionar campos `permissions []string` e `roles []string` ao `auth.Claims`
- [ ] `GenerateAccessToken` resolve permissions via roles do admin no banco
- [ ] Backward-compat: `superadmin=true` continua bypassando todas as checks

### Middleware
- [ ] `RequirePermission(perms ...string)` — verifica ao menos uma permission no claim
- [ ] `RequireRole(roles ...string)` — verifica ao menos um role no claim
- [ ] Feature flag `rbac.enabled` — quando false, middleware é no-op
- [ ] Domain scoping no middleware: cruza permission com `Domains` do claim

### API REST (`/api/v1/rbac`)
- [ ] `GET  /rbac/roles` — lista roles
- [ ] `POST /rbac/roles` — cria role customizado
- [ ] `GET  /rbac/roles/:id` — detalhe com permissions
- [ ] `PUT  /rbac/roles/:id` — edita role (bloqueia system=true)
- [ ] `DELETE /rbac/roles/:id` — remove role (bloqueia system=true)
- [ ] `GET  /rbac/permissions` — lista catálogo de permissions
- [ ] `GET  /rbac/admins/:username/roles` — roles do admin
- [ ] `POST /rbac/admins/:username/roles` — atribui role
- [ ] `DELETE /rbac/admins/:username/roles/:roleId` — remove role

### Integração nas Rotas Existentes
- [ ] Aplicar `RequirePermission` em todos os endpoints de `/api/v1`
- [ ] Endpoints de escrita exigem permission `<resource>:write`
- [ ] Endpoints de deleção exigem permission `<resource>:delete`
- [ ] Endpoints de leitura exigem permission `<resource>:read`
- [ ] `/admins/*` e `/transports/*` exigem adicionalmente `superadmin` ou role equivalente

### CLI
- [ ] Subcomando `migrate rbac` — cria tabelas e executa seed
- [ ] Subcomando `rbac assign <username> <role> [domain]` — atribui role via CLI

### Frontend (Vue 3)
- [ ] Página de gerenciamento de roles (`/settings/roles`)
- [ ] Componente de atribuição de roles no formulário de edição de admin
- [ ] Guards de rota Vue baseados em permissions do JWT

## Non-Goals

- ABAC (Attribute-Based Access Control) — fora do escopo desta feature
- Roles hierárquicos (herança de permissions entre roles)
- UI de gerenciamento de permissions individuais (apenas via seed/code)
- Suporte a múltiplos tenants além do domain scoping existente
