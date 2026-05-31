# Features — api-rest

## Implementado

- [x] Auth JWT: login, refresh, logout, me
- [x] Domains — CRUD completo (scoped por admin)
- [x] Mailboxes — CRUD completo (scoped por admin)
- [x] Aliases — CRUD completo (scoped por admin)
- [x] Alias Domains — CRUD completo (scoped por admin)
- [x] Admins — CRUD completo (superadmin)
- [x] Transports — CRUD completo (superadmin)
- [x] API Keys — CRUD (`/settings/apikeys`)
- [x] User self-service: perfil, forwarding, senha, férias/vacation
- [x] Dashboard stats
- [x] Logs paginados (admin log + maillog)
- [x] OpenAPI/Swagger embutido

## Pendente / Planejado

- [ ] Endpoint de broadcast email (`/mailboxes/broadcast` ou similar)
- [ ] Paginação padronizada com cursor (atualmente por offset)
- [ ] Filtros e busca nos endpoints de listagem
- [ ] Rate limiting por usuário (além do global atual)
- [ ] Webhook / evento de criação de domínio/mailbox

## Non-Goals

- Suporte a múltiplos backends de banco (apenas MariaDB/PostgreSQL via config)
- API GraphQL
- Compatibilidade com API do PostfixAdmin PHP original
